package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/nullstone.v0/app"
)

var Deploy = func(providers app.Providers) cli.Command {
	return cli.Command{
		Name:      "deploy",
		Usage:     "Deploy application",
		UsageText: "nullstone deploy <app-name> <env-name> [options]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: "stack",
				Usage: `The stack name where the app resides.
       This is only required if multiple apps have the same 'app-name'.`,
			},
			cli.StringFlag{
				Name: "version",
				Usage: `Update the application version.
       app/container: The docker image tag will be set to the version. If a version is not specified, the service will be redeployed with existing configuration.
       app/serverless: The version of the artifact uploaded during 'push'. Version is required to use deploy command.`,
			},
		},
		Action: func(c *cli.Context) error {
			_, cfg, err := SetupProfileCmd(c)
			if err != nil {
				return err
			}

			if c.NArg() != 2 {
				cli.ShowCommandHelp(c, "deploy")
				return fmt.Errorf("invalid usage")
			}
			appName := c.Args().Get(0)
			envName := c.Args().Get(1)
			userConfig := map[string]string{
				"version": c.String("version"),
			}

			finder := NsFinder{Config: cfg}
			app, workspace, err := finder.GetAppAndWorkspace(appName, c.String("stack-name"), envName)
			if err != nil {
				return err
			}

			provider := providers.Find(workspace.Module.Category, workspace.Module.Type)
			if provider == nil {
				return fmt.Errorf("unable to deploy, this CLI does not support category=%s, type=%s", workspace.Module.Category, workspace.Module.Type)
			}
			return provider.Deploy(cfg, app, workspace, userConfig)
		},
	}
}
