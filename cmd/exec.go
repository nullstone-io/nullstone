package cmd

import (
	"context"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"log"
	"os"
)

var (
	logger = log.New(os.Stderr, "", 0)
)

var Exec = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "exec",
		Usage:     "Execute command on running service. Defaults command to '/bin/sh' which acts as opening a shell to the running container.",
		UsageText: "nullstone exec [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options] [command]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvFlag,
			TaskFlag,
		},
		Action: func(c *cli.Context) error {
			userConfig := map[string]string{
				"task": c.String("task"),
				"cmd":  "/bin/sh",
			}
			if c.Args().Len() >= 1 {
				userConfig["cmd"] = c.Args().Get(c.Args().Len() - 1)
			}

			return AppEnvAction(c, providers, func(ctx context.Context, cfg api.Config, provider app.Provider, details app.Details) error {
				return provider.Exec(ctx, logger, cfg, details, userConfig)
			})
		},
	}
}
