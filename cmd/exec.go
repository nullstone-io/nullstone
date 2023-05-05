package cmd

import (
	"context"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

var Exec = func(providers admin.Providers) *cli.Command {
	return &cli.Command{
		Name:        "exec",
		Description: "Executes a command on a container or the virtual machine for the given application. Defaults command to '/bin/sh' which acts as opening a shell to the running container or virtual machine.",
		Usage:       "Execute a command on running service",
		UsageText:   "nullstone exec [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options] [command]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvFlag,
			TaskFlag,
			ReplicaFlag,
			ContainerFlag,
		},
		Action: func(c *cli.Context) error {
			cmd := "/bin/sh"
			if c.Args().Len() >= 1 {
				cmd = c.Args().Get(c.Args().Len() - 1)
			}

			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				remoter, err := providers.FindRemoter(logging.StandardOsWriters{}, cfg, appDetails)
				if err != nil {
					return err
				}
				options := admin.RemoteOptions{
					Task:      c.String("task"),
					Replica:   c.String("replica"),
					Container: c.String("container"),
				}
				return remoter.Exec(ctx, options, cmd)
			})
		},
	}
}
