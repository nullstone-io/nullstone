package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
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
					return fmt.Errorf("launch is not supported for this app")
				}

				err := push(ctx, cfg, appDetails, osWriters, factory, source, version)
				if err != nil {
					return err
				}
				reference, err := deploy(ctx, cfg, appDetails, osWriters, factory, version)
				if err != nil {
					return err
				}
				if err := waitHealthy(ctx, cfg, appDetails, osWriters, factory, reference); err != nil {
					return err
				}
				return nil
			})
		},
	}
}
