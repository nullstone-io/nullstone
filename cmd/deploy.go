package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/aws/fargate"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"gopkg.in/nullstone-io/nullstone.v0/deploy"
)

var (
	DefaultDeployers = deploy.Deployers{
		fargate.Deployer{},
		// TODO: Add support for other app categories
		// TODO: Add support for other providers
	}
)

var Deploy = cli.Command{
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
		profile, err := config.LoadProfile(GetProfile(c))
		if err != nil {
			return err
		}

		if c.NArg() != 2 {
			return cli.ShowCommandHelp(c, "deploy")
		}
		appName := c.Args().Get(0)
		envName := c.Args().Get(1)
		userConfig := map[string]string{
			"imageTag": c.String("image-tag"),
		}

		config := api.DefaultConfig()
		config.BaseAddress = profile.Address
		config.ApiKey = profile.ApiKey
		config.OrgName = GetOrg(c, *profile)
		if config.OrgName == "" {
			return ErrMissingOrg
		}
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

		return DefaultDeployers.Deploy(config, app, workspace, userConfig)
	},
}
