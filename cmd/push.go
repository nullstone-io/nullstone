package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
)

// Push command performs a docker push to an authenticated image registry configured against an app/container
var Push = func(providers app.Providers) cli.Command {
	return cli.Command{
		Name:      "push",
		Usage:     "Push artifact",
		UsageText: "nullstone push <app-name> <env-name> [options]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:     "source",
				Usage:    "The source docker image to push. This follows the same syntax as `docker push NAME[:TAG]`.",
				Required: true,
			},
			cli.StringFlag{
				Name:  "image-tag",
				Usage: "Push the image with this tag instead of the source. If not specified, will use the source tag.",
				Value: "latest",
			},
		},
		Action: func(c *cli.Context) error {
			_, cfg, err := SetupProfileCmd(c)
			if err != nil {
				return err
			}
			client := api.Client{Config: cfg}

			if c.NArg() != 2 {
				return cli.ShowCommandHelp(c, "push")
			}
			appName := c.Args().Get(0)
			envName := c.Args().Get(1)
			userConfig := map[string]string{
				"source":   c.String("source"),
				"imageTag": c.String("image-tag"),
			}

			app, err := client.Apps().Get(appName)
			if err != nil {
				return fmt.Errorf("error retrieving application %q: %w", appName, err)
			} else if app == nil {
				return fmt.Errorf("application %q does not exist", appName)
			}

			workspace, err := client.Workspaces().Get(app.StackName, app.Block.Name, envName)
			if err != nil {
				return fmt.Errorf("error retrieving workspace: %w", err)
			} else if workspace == nil {
				return fmt.Errorf("workspace %q does not exist", err)
			}
			if workspace.Status != types.WorkspaceStatusProvisioned {
				return fmt.Errorf("app %q has not been provisioned in %q environment yet", app.Name, workspace.EnvName)
			}
			if workspace.Module == nil {
				return fmt.Errorf("unknown module for workspace, cannot perform deployment")
			}

			provider := providers.Find(workspace.Module.Category, workspace.Module.Type)
			if provider == nil {
				return fmt.Errorf("unable to push, this CLI does not support category=%s, type=%s", workspace.Module.Category, workspace.Module.Type)
			}

			return provider.Push(cfg, app, workspace, userConfig)
		},
	}
}
