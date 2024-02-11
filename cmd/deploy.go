package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/go-api-client.v0/ws"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
	version2 "gopkg.in/nullstone-io/nullstone.v0/version"
	"os"
	"time"
)

var Deploy = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:        "deploy",
		Description: "Deploy a new version of your code for this application. This command works in tandem with the `nullstone push` command. This command deploys the artifacts that were uploaded during the `push` command.",
		Usage:       "Deploy application",
		UsageText:   "nullstone deploy [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppVersionFlag,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait for the deploy to complete and stream the logs to the console.",
			},
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				version, wait := c.String("version"), c.IsSet("wait")

				if version == "" {
					fmt.Fprintf(os.Stderr, "No version specified. Defaulting version based on current git commit sha...\n")
					pusher, err := getPusher(providers, cfg, appDetails)
					if err != nil {
						return err
					}

					version, err = getCurrentVersion(ctx, *pusher)
					if err != nil {
						return err
					}
					fmt.Fprintf(os.Stderr, "Version defaulted to: %s\n", version)
				}

				fmt.Fprintln(os.Stderr, "Creating deploy...")
				deploy, err := CreateDeploy(cfg, appDetails, version)
				if err != nil {
					return err
				}

				fmt.Fprintln(os.Stderr)
				return streamDeployLogs(ctx, cfg, *deploy, wait)
			})
		},
	}
}

func getCurrentVersion(ctx context.Context, pusher app.Pusher) (string, error) {
	shortSha, err := vcs.GetCurrentShortCommitSha()
	if err != nil {
		return "", fmt.Errorf("error calculating version: %w", err)
	}

	artifacts, err := pusher.ListArtifacts(ctx)
	if err != nil {
		return "", fmt.Errorf("error calculating version: %w", err)
	}

	seq := version2.FindLatestVersionSequence(shortSha, artifacts)
	if err != nil {
		return "", fmt.Errorf("error calculating version: %w", err)
	}

	if seq == 0 {
		return "", fmt.Errorf("no artifacts found for this commit SHA (%s) - you must perform a successful push before deploying", shortSha)
	}
	return fmt.Sprintf("%s-%d", shortSha, seq), nil
}

func streamDeployLogs(ctx context.Context, cfg api.Config, deploy types.Deploy, wait bool) error {
	fmt.Fprintln(os.Stderr, "Waiting for deploy logs...")
	client := api.Client{Config: cfg}
	msgs, err := client.DeployLogs().Watch(ctx, deploy.StackId, deploy.Id, ws.RetryInfinite(2*time.Second))
	if err != nil {
		return fmt.Errorf("error connecting to deploy logs: %w", err)
	}
	for msg := range msgs {
		if msg.Type == "error" {
			return fmt.Errorf(msg.Content)
		}
		if !wait && msg.Context == types.DeployPhaseWaitHealthy {
			// Stop streaming logs if we receive a log message from wait-healthy and no --wait
			break
		}
		fmt.Fprint(os.Stderr, msg.Content)
	}

	updated, err := client.Deploys().Get(deploy.StackId, deploy.AppId, deploy.EnvId, deploy.Id)
	if err != nil {
		return fmt.Errorf("error retrieving deploy status: %w", err)
	}
	fmt.Fprintln(os.Stdout, updated.Reference)
	switch updated.Status {
	case types.DeployStatusCancelled:
		return fmt.Errorf("Deploy was cancelled.")
	case types.DeployStatusFailed:
		return fmt.Errorf("Deploy failed to complete: %s", updated.StatusMessage)
	}
	return nil
}
