package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"gopkg.in/nullstone-io/nullstone.v0/deploys"
)

// Launch command performs push, deploy, and logs
var Launch = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "launch",
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

				deploy, err := CreateDeploy(cfg, appDetails, version)
				if err != nil {
					return err
				}
				return deploys.StreamLogs(ctx, cfg, deploy)
			})
		},
	}
}
