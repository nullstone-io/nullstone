package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"os"
)

var Run = func(appProviders app.Providers, providers admin.Providers) *cli.Command {
	return &cli.Command{
		Name:        "Run",
		Description: "Starts a new container/serverless for the given Nullstone job/task. ",
		Usage:       "Starts a new job/task",
		UsageText:   "nullstone run [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options] [command]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvFlag,
			ContainerFlag,
		},
		Action: func(c *cli.Context) error {
			var cmd []string
			if c.Args().Present() {
				cmd = c.Args().Slice()
			}

			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				client := api.Client{Config: cfg}
				user, err := client.CurrentUser().Get(ctx)
				if err != nil {
					return fmt.Errorf("unable to fetch the current user")
				}
				if user == nil {
					return fmt.Errorf("unable to load the current user info")
				}

				source := outputs.ApiRetrieverSource{Config: cfg}

				logStreamer, err := appProviders.FindLogStreamer(ctx, logging.StandardOsWriters{}, source, appDetails)
				if err != nil {
					return err
				}

				remoter, err := providers.FindRemoter(ctx, logging.StandardOsWriters{}, source, appDetails)
				if err != nil {
					return err
				}
				options := admin.RunOptions{
					Container:   c.String("container"),
					Username:    user.Name,
					LogStreamer: logStreamer,
					LogEmitter:  app.NewWriterLogEmitter(os.Stdout),
				}
				return remoter.Run(ctx, options, cmd)
			})
		},
	}
}
