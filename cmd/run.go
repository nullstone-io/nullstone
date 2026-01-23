package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

var RunEnvVarFlag = &cli.StringSliceFlag{
	Name:     "env",
	Aliases:  []string{"e"},
	Usage:    "Environment variable to pass to the job/task",
	Required: false,
}

var Run = func(appProviders app.Providers, providers admin.Providers) *cli.Command {
	return &cli.Command{
		Name:        "run",
		Description: "Starts a new container/serverless for the given Nullstone job/task. ",
		Usage:       "Starts a new job/task",
		UsageText:   "nullstone run [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options] [command]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvFlag,
			ContainerFlag,
			RunEnvVarFlag,
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

				rawEnvVars := c.StringSlice(RunEnvVarFlag.Name)
				var envVars map[string]string
				for _, raw := range rawEnvVars {
					before, after, ok := strings.Cut(raw, "=")
					if !ok {
						return fmt.Errorf("invalid --env flag, expected <NAME>=<value>")
					}
					envVars[before] = after
				}

				source := outputs.ApiRetrieverSource{Config: cfg}
				osWriters := logging.StandardOsWriters{}

				logStreamer, err := appProviders.FindLogStreamer(ctx, osWriters, source, appDetails)
				if err != nil {
					return err
				}

				remoter, err := providers.FindRemoter(ctx, osWriters, source, appDetails)
				if err != nil {
					return err
				}
				options := admin.RunOptions{
					Container:   c.String(ContainerFlag.Name),
					Username:    user.Name,
					LogStreamer: logStreamer,
					LogEmitter:  app.NewWriterLogEmitter(os.Stdout),
				}
				if remoter == nil {
					return fmt.Errorf("run is not supported for this workspace")
				}
				return remoter.Run(ctx, options, cmd, envVars)
			})
		},
	}
}
