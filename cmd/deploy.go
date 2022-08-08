package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/nullstone.v0/deploys"
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
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
			},
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				version, wait := DetectAppVersion(c), c.IsSet("wait")
				if version == "" {
					return fmt.Errorf("no version specified, version is required to create a deploy")
				}

				deploy, err := CreateDeploy(cfg, appDetails, version)
				if err != nil {
					return err
				}

				if wait {
					// TODO: We should always stream logs, but if --wait is not specified, we would skip "wait-healthy" phase
					return deploys.StreamLogs(ctx, cfg, deploy)
				}
				return nil
			})
		},
	}
}
