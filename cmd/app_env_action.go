package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"log"
	"os"
)

type AppEnvActionFn func(ctx context.Context, cfg api.Config, provider app.Provider, details app.Details) error

func AppEnvAction(c *cli.Context, providers app.Providers, fn AppEnvActionFn) error {
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

		provider := providers.Find(*appDetails.Module)
		if provider == nil {
			return fmt.Errorf("this CLI does not support application category=%s, type=%s", appDetails.Module.Category, appDetails.Module.Type)
		}

		return CancellableAction(func(ctx context.Context) error {
			return fn(ctx, cfg, provider, appDetails)
		})
	})
}
