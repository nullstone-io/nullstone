package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"log"
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
			return AppEnvAction(c, providers, func(ctx context.Context, cfg api.Config, provider app.Provider, details app.Details) error {
				deployer, err := provider.NewDeployer()
				if err != nil {
					return fmt.Errorf("Unable to identify infrastructure: %w", err)
				}
				reference, err := deployer.Deploy(ctx, DetectAppVersion(c))
				if err != nil {
					return err
				} else if reference != nil {
					log.Printf("Deployment ID: %s\n", *reference)
				}
				return nil
			})
		},
	}
}
