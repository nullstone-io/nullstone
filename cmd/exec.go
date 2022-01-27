package cmd

import (
	"context"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
)

var Exec = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "exec",
		Usage:     "Execute command on running service. Defaults command to '/bin/sh' which acts as opening a shell to the running container.",
		UsageText: "nullstone exec [options] <app-name> <env-name> [command]",
		Flags: []cli.Flag{
			StackFlag,
			TaskFlag,
		},
		Action: func(c *cli.Context) error {
			return AppEnvAction(c, providers, func(ctx context.Context, cfg api.Config, provider app.Provider, details app.Details) error {
				userConfig := map[string]string{
					"task": c.String("task"),
					"cmd":  "/bin/sh",
				}
				if c.Args().Len() > 2 {
					userConfig["cmd"] = c.Args().Get(2)
				}
				return provider.Exec(cfg, details, userConfig)
			})
		},
	}
}
