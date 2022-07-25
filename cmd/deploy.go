package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

var Deploy = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "deploy",
		Usage:     "Deploy application",
		UsageText: "nullstone deploy [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppVersionFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				osWriters := logging.StandardOsWriters{}
				deployer, err := providers.FindDeployer(osWriters, cfg, appDetails)
				if err != nil {
					return fmt.Errorf("error creating app deployer: %w", err)
				} else if deployer == nil {
					return fmt.Errorf("this CLI does not support application category=%s, type=%s", appDetails.Module.Category, appDetails.Module.Type)
				}
				reference, err := deployer.Deploy(ctx, DetectAppVersion(c))
				if err != nil {
					return err
				}

				fmt.Fprintln(osWriters.Stdout(), reference)
				return nil
			})
		},
	}
}
