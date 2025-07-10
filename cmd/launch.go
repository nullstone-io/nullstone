package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	version2 "gopkg.in/nullstone-io/nullstone.v0/version"
)

// Launch command performs push, deploy, and logs
var Launch = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name: "launch",
		Description: "This command will first upload (push) an artifact containing the source for your application. Then it will deploy it to the given environment and tail the logs for the deployment." +
			"This command is the same as running `nullstone push` followed by `nullstone deploy -w`.",
		Usage:     "Launch application (push + deploy + wait-healthy)",
		UsageText: "nullstone launch [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppSourceFlag,
			AppVersionFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				osWriters := CliOsWriters{Context: c}
				stderr := osWriters.Stderr()
				source, version := c.String("source"), c.String("version")

				pusher, err := getPusher(providers, cfg, appDetails)
				if err != nil {
					return err
				}

				commitSha := ""
				if version == "" {
					fmt.Fprintf(stderr, "No version specified. Defaulting version based on current git commit sha...\n")
					info, err := version2.CalcNew(ctx, pusher)
					if err != nil {
						return err
					}
					version = info.Version
					commitSha = info.CommitSha
					fmt.Fprintf(stderr, "Version defaulted to: %s\n", version)
				}

				if err := recordArtifact(ctx, osWriters, cfg, appDetails, version); err != nil {
					return err
				}

				err = push(ctx, osWriters, pusher, source, version)
				if err != nil {
					return err
				}

				fmt.Fprintln(stderr, "Creating deploy...")
				result, err := CreateDeploy(cfg, appDetails, commitSha, version)
				if err != nil {
					return err
				}

				fmt.Fprintln(stderr)
				if result.Deploy != nil {
					return streamDeployLogs(ctx, osWriters, cfg, *result.Deploy, true)
				} else if result.IntentWorkflow != nil {
					return streamDeployIntentLogs(ctx, osWriters, cfg, appDetails, *result.IntentWorkflow, true)
				}
				fmt.Fprintln(stderr, "Unable to stream deployment logs")
				return nil
			})
		},
	}
}
