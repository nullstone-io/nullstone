package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"os"
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

				err = push(ctx, *pusher, source, version)
				if err != nil {
					return err
				}

				fmt.Fprintln(os.Stderr, "Creating deploy...")
				deploy, err := CreateDeploy(cfg, appDetails, version)
				if err != nil {
					return err
				}

				fmt.Fprintln(os.Stderr)
				return streamDeployLogs(ctx, cfg, *deploy, true)
			})
		},
	}
}
