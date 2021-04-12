package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/aws/fargate"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"os"
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
		imageTag := c.String("image-tag")

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

		if workspace.Status != types.WorkspaceStatusProvisioned {
			return fmt.Errorf("app %q has not been provisioned in %q environment yet", appName, envName)
		}

		if workspace.Module == nil {
			return fmt.Errorf("unknown module for workspace, cannot perform deployment")
		}

		fmt.Fprintf(os.Stderr, "Deploying app %q\n", appName)
		type deployer interface {
			Deploy(workspace *types.Workspace, imageTag string) error
		}
		var deployers = map[types.CategoryName]deployer{
			types.CategoryAppContainer: fargate.Deployer{},
			// TODO: Add support for other app categories
		}
		// TODO: Assumes AWS, add support for other providers

		dep, ok := deployers[workspace.Module.Category]
		if !ok {
			return fmt.Errorf("unknown deployment pattern %s", workspace.Module.Category)
		}

		if err := dep.Deploy(workspace, imageTag); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Deployed app %q\n", app.Name)
		return nil
	},
}
