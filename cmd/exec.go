package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cristalhq/jwt/v3"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"os"
)

var Exec = func(appProviders app.Providers, providers admin.Providers) *cli.Command {
	return &cli.Command{
		Name:        "exec",
		Description: "Executes a command on a container or the virtual machine for the given application. Defaults command to '/bin/sh' which acts as opening a shell to the running container or virtual machine.",
		Usage:       "Execute a command on running service",
		UsageText:   "nullstone exec [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options] [command]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvFlag,
			InstanceFlag,
			TaskFlag,
			PodFlag,
			ContainerFlag,
			WaitForLaunchFlag,
		},
		Action: func(c *cli.Context) error {
			var cmd []string
			if c.Args().Present() {
				cmd = c.Args().Slice()
			}

			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				client := api.Client{Config: cfg}
				user, err := client.CurrentUser().Get()
				if err != nil {
					return fmt.Errorf("unable to fetch the current user")
				}
				if user == nil {
					return fmt.Errorf("unable to load the current user info")
				}

				source := outputs.ApiRetrieverSource{Config: cfg}

				logStreamer, err := appProviders.FindLogStreamer(logging.StandardOsWriters{}, source, appDetails)
				if err != nil {
					return err
				}

				remoter, err := providers.FindRemoter(logging.StandardOsWriters{}, source, appDetails)
				if err != nil {
					return err
				}
				options := admin.RemoteOptions{
					Instance:    c.String("instance"),
					Task:        c.String("task"),
					Pod:         c.String("pod"),
					Container:   c.String("container"),
					Username:    user.Name,
					LogStreamer: logStreamer,
					LogEmitter:  app.NewWriterLogEmitter(os.Stdout),
				}
				return remoter.Exec(ctx, options, cmd)
			})
		},
	}
}

type Claims struct {
	jwt.StandardClaims
	Email    string            `json:"email"`
	Picture  string            `json:"picture"`
	Username string            `json:"https://nullstone.io/username"`
	Roles    map[string]string `json:"https://nullstone.io/roles"`
}

func getClaims(rawToken string) (*Claims, error) {
	token, err := jwt.ParseString(rawToken)
	if err != nil {
		return nil, err
	}
	var claims Claims
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}
