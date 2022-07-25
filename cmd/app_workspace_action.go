package cmd

import (
	"context"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"log"
	"os"
)

type AppWorkspaceFn func(ctx context.Context, cfg api.Config, appDetails app.Details) error

func AppWorkspaceAction(c *cli.Context, fn AppWorkspaceFn) error {
	_, cfg, err := SetupProfileCmd(c)
	if err != nil {
		return err
	}

	return ParseAppEnv(c, true, func(stackName, appName, envName string) error {
		logger := log.New(os.Stderr, "", 0)
		logger.Printf("Performing application command (Org=%s, App=%s, Stack=%s, Env=%s)", cfg.OrgName, appName, stackName, envName)
		logger.Println()

		finder := NsFinder{Config: cfg}
		appDetails, err := finder.FindAppDetails(appName, stackName, envName)
		if err != nil {
			return err
		}

		return CancellableAction(func(ctx context.Context) error {
			return fn(ctx, cfg, app.Details{
				App:       appDetails.App,
				Env:       appDetails.Env,
				Workspace: appDetails.Workspace,
				Module:    appDetails.Module,
			})
		})
	})
}
