package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
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
				source, version := c.String("source"), DetectAppVersion(c)
				osWriters := logging.StandardOsWriters{}
				factory := providers.FindFactory(*appDetails.Module)
				if factory == nil {
					return fmt.Errorf("this app module is not supported")
				}

				err := push(ctx, cfg, appDetails, osWriters, factory, source, version)
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
