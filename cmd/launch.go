package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/app_logs"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"log"
	"os"
)

// Launch command performs push, deploy, and logs
var Launch = func(providers app.Providers, logProviders app_logs.Providers) *cli.Command {
	return &cli.Command{
		Name:      "launch",
		Usage:     "Launch application (push, deploy, log)",
		UsageText: "nullstone launch [options] <app-name> <env-name>",
		Flags: []cli.Flag{
			StackFlag,
			AppSourceFlag,
			AppVersionFlag,
		},
		Action: func(c *cli.Context) error {
			return AppAction(c, providers, func(ctx context.Context, cfg api.Config, provider app.Provider, details app.Details) error {
				logger := log.New(os.Stderr, "", 0)

				userConfig := map[string]string{
					"source":  c.String("source"),
					"version": c.String("version"),
				}

				logger.Println("Pushing app artifact...")
				if err := provider.Push(cfg, details, userConfig); err != nil {
					return fmt.Errorf("error pushing artifact: %w", err)
				}
				logger.Println()

				logger.Println("Deploying application...")
				if err := provider.Deploy(cfg, details, userConfig); err != nil {
					return fmt.Errorf("error deploying app: %w", err)
				}
				logger.Println()

				logger.Println("Tailing application logs...")
				logProvider, err := logProviders.Identify(provider.DefaultLogProvider(), cfg, details)
				if err != nil {
					return err
				}
				return logProvider.Stream(ctx, cfg, details, config.LogStreamOptions{Out: os.Stdout})
			})
		},
	}
}
