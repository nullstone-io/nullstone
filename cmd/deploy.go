package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"time"
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
				deployer, err := providers.FindDeployer(logging.StandardOsWriters{}, cfg, appDetails)
				if err != nil {
					return fmt.Errorf("error creating app deployer: %w", err)
				} else if deployer == nil {
					return fmt.Errorf("this CLI does not support application category=%s, type=%s", appDetails.Module.Category, appDetails.Module.Type)
				}
				reference, err := deployer.Deploy(ctx, DetectAppVersion(c))
				if err != nil {
					return err
				}

				deployStatusGetter, err := providers.FindDeployStatusGetter(logging.StandardOsWriters{}, cfg, appDetails)
				if err != nil {
					return fmt.Errorf("error creating app deployment status analyzer: %w", err)
				} else if deployStatusGetter == nil {
					// If we don't have a way of retrieving status, we are just going to complete the deployment
					return nil
				}

				for {
					status, err := deployStatusGetter.GetDeployStatus(ctx, reference)
					if err != nil {
						return fmt.Errorf("error querying app deployment status: %w", err)
					}
					switch status {
					case app.RolloutStatusComplete:
						return nil
					case app.RolloutStatusInProgress:
						return nil
					case app.RolloutStatusFailed:
						return fmt.Errorf("deployment failed")
					case app.RolloutStatusUnknown:
						return fmt.Errorf("unknown app deployment status")
					}

					select {
					case <-ctx.Done():
						return fmt.Errorf("cancelled")
					case <-time.After(5 * time.Second):
					}
				}
			})
		},
	}
}
