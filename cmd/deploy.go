package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/aws/fargate"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

var Deploy = cli.Command{
	Name:      "deploy",
	Usage:     "Deploy application",
	UsageText: "nullstone deploy <app-name> <env-name> [options]",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "image-tag",
			Usage: "",
		},
	},
	Action: func(c *cli.Context) error {
		profile, err := config.LoadProfile(GetProfile(c.Parent()))
		if err != nil {
			return err
		}

		if c.NArg() != 2 {
			return fmt.Errorf("invalid number of arguments, expected 2, got %d", c.NArg())
		}
		appName := c.Args().Get(0)
		envName := c.Args().Get(1)

		config := api.DefaultConfig()
		config.BaseAddress = profile.Address
		config.ApiKey = profile.ApiKey
		config.OrgName = GetOrg(c, *profile)
		client := api.Client{Config: config}

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

		if workspace.Module == nil {
			return fmt.Errorf("unknown module for workspace, cannot perform deployment")
		}

		type deployerFn func(app *types.Application, workspace *types.Workspace) error
		var deployers = map[types.CategoryName]deployerFn{
			types.CategoryAppContainer: fargate.DeployContainer,
			// TODO: Add support for other app categories
		}
		// TODO: Assumes AWS, add support for other providers

		fn, ok := deployers[workspace.Module.Category]
		if !ok {
			return fmt.Errorf("unknown deployment pattern %s", workspace.Module.Category)
		}

		return fn(app, workspace)
	},
}
