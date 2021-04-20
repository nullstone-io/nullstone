package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
)

var Deploy = func(providers app.Providers) cli.Command {
	return cli.Command{
		Name:      "deploy",
		Usage:     "Deploy application",
		UsageText: "nullstone deploy <app-name> <env-name> [options]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "image-tag",
				Usage: "Update the docker image tag for apps defined as 'app/container'. If not specified, will force a deployment.",
			},
		},
		Action: func(c *cli.Context) error {
			_, cfg, err := SetupProfileCmd(c)
			if err != nil {
				return err
			}
			client := api.Client{Config: cfg}

			if c.NArg() != 2 {
				return cli.ShowCommandHelp(c, "deploy")
			}
			appName := c.Args().Get(0)
			envName := c.Args().Get(1)
			userConfig := map[string]string{
				"version": c.String("version"),
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
				return fmt.Errorf("unable to deploy, this CLI does not support category=%s, type=%s", workspace.Module.Category, workspace.Module.Type)
			}
			return provider.Deploy(cfg, app, workspace, userConfig)
		},
	}
}
