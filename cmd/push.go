package cmd

import (
	"context"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
)

// Push command performs a docker push to an authenticated image registry configured against an app/container
var Push = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "push",
		Usage:     "Push artifact",
		UsageText: "nullstone push [options] <app-name> <env-name>",
		Flags: []cli.Flag{
			StackFlag,
			AppSourceFlag,
			AppVersionFlag,
		},
		Action: func(c *cli.Context) error {
			return AppEnvAction(c, providers, func(ctx context.Context, cfg api.Config, provider app.Provider, details app.Details) error {
				userConfig := map[string]string{
					"source":  c.String("source"),
					"version": c.String("version"),
				}
				return provider.Push(cfg, details, userConfig)
			})
		},
	}
}
