package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
	version2 "gopkg.in/nullstone-io/nullstone.v0/version"
	"os"
)

// Push command performs a docker push to an authenticated image registry configured against an app/container
var Push = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:        "push",
		Description: "Upload (push) an artifact containing the source for your application. Specify a semver version to associate with the artifact. The version specified can be used in the deploy command to select this artifact.",
		Usage:       "Push artifact",
		UsageText:   "nullstone push [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppSourceFlag,
			AppVersionFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				source, version := c.String("source"), c.String("version")

				pusher, err := getPusher(providers, cfg, appDetails)
				if err != nil {
					return err
				}

				if version == "" {
					fmt.Fprintf(os.Stderr, "No version specified. Defaulting version based on current git commit sha...\n")
					version, err = calcNewVersion(ctx, *pusher)
					if err != nil {
						return err
					}
					fmt.Fprintf(os.Stderr, "Version defaulted to: %s\n", version)
				}

				return push(ctx, *pusher, source, version)
			})
		},
	}
}

func getPusher(providers app.Providers, cfg api.Config, appDetails app.Details) (*app.Pusher, error) {
	osWriters := logging.StandardOsWriters{}
	provider := providers.FindFactory(*appDetails.Module)
	if provider == nil {
		return nil, fmt.Errorf("push is not supported for this app")
	}

	if provider.NewPusher == nil {
		return nil, fmt.Errorf("this app does not support push")
	}
	retriever := outputs.ApiRetrieverSource{Config: cfg}
	pusher, err := provider.NewPusher(osWriters, retriever, appDetails)
	if err != nil {
		return nil, fmt.Errorf("error creating app pusher: %w", err)
	} else if pusher == nil {
		return nil, fmt.Errorf("this application category=%s, type=%s does not support push", appDetails.Module.Category, appDetails.Module.Type)
	}
	return &pusher, nil
}

func calcNewVersion(ctx context.Context, pusher app.Pusher) (string, error) {
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

	version := fmt.Sprintf("%s-%d", shortSha, seq+1)

	return version, nil
}

func push(ctx context.Context, pusher app.Pusher, source, version string) error {
	fmt.Fprintln(os.Stderr, "Pushing app artifact...")
	if err := pusher.Push(ctx, source, version); err != nil {
		return fmt.Errorf("error pushing artifact: %w", err)
	}
	fmt.Fprintln(os.Stderr, "App artifact pushed.")
	fmt.Fprintln(os.Stderr, "")
	return nil
}
