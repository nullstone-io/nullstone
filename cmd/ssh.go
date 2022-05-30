package cmd

import (
	"context"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
)

var Ssh = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "ssh",
		Usage:     "SSH into a running service. Use to forward ports from remote service or hosts.",
		UsageText: "nullstone ssh [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvFlag,
			TaskFlag,
			&cli.StringSliceFlag{
				Name:    "forward",
				Aliases: []string{"L"},
				Usage:   "Use this to forward ports from host to local machine. Format: <local-port>:[<remote-host>]:<remote-port>",
			},
		},
		Action: func(c *cli.Context) error {
			userConfig := map[string]any{
				"task": c.String("task"),
			}

			return AppEnvAction(c, providers, func(ctx context.Context, cfg api.Config, provider app.Provider, details app.Details) error {
				return provider.Ssh(ctx, cfg, details, userConfig)
			})
		},
	}
}
