package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/config"
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

			forwards := make([]config.PortForward, 0)
			for _, arg := range c.StringSlice("forward") {
				pf, err := config.ParsePortForward(arg)
				if err != nil {
					return fmt.Errorf("invalid format for --forward/-L: %w", err)
				}
				forwards = append(forwards, pf)
			}
			userConfig["forwards"] = forwards

			return AppEnvAction(c, providers, func(ctx context.Context, cfg api.Config, provider app.ProviderOld, details app.Details) error {
				return provider.Ssh(ctx, cfg, details, userConfig)
			})
		},
	}
}
