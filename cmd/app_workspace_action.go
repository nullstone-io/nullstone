package cmd

import (
	"context"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"log"
)

type AppWorkspaceFn func(ctx context.Context, cfg api.Config, appDetails app.Details) error

func AppWorkspaceAction(c *cli.Context, fn AppWorkspaceFn) error {
	_, cfg, err := SetupProfileCmd(c)
	if err != nil {
		return err
	}

	waitForLaunch := c.Bool(WaitForLaunchFlag.Name)
	osWriters := CliOsWriters{Context: c}

	return ParseAppEnv(c, true, func(stackName, appName, envName string) error {
		logger := log.New(osWriters.Stderr(), "", 0)
		logger.Printf("Performing application command (Org=%s, App=%s, Stack=%s, Env=%s)", cfg.OrgName, appName, stackName, envName)
		logger.Println()

		appDetails, err := FindAppDetails(cfg, appName, stackName, envName)
		if err != nil {
			return err
		}

		return CancellableAction(func(ctx context.Context) error {
			if err := WaitForLaunch(ctx, osWriters, cfg, appDetails, waitForLaunch); err != nil {
				return err
			}

			return fn(ctx, cfg, app.Details{
				App:       appDetails.App,
				Env:       appDetails.Env,
				Workspace: appDetails.Workspace,
				Module:    appDetails.Module,
			})
		})
	})
}
