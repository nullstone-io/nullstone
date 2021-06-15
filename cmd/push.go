package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/nullstone.v0/app"
)

// Push command performs a docker push to an authenticated image registry configured against an app/container
var Push = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "push",
		Usage:     "Push artifact",
		UsageText: "nullstone push <app-name> <env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppSourceFlag,
			&cli.StringFlag{
				Name: "version",
				Usage: `Push the artifact with this version.
       app/container: If specified, will push the docker image with version as the image tag. Otherwise, uses source tag.
       app/serverless: This is required to upload the artifact.`,
			},
		},
		Action: func(c *cli.Context) error {
			_, cfg, err := SetupProfileCmd(c)
			if err != nil {
				return err
			}

			if c.NArg() != 2 {
				cli.ShowCommandHelp(c, "push")
				return fmt.Errorf("invalid usage")
			}
			appName := c.Args().Get(0)
			envName := c.Args().Get(1)
			userConfig := map[string]string{
				"source":  c.String("source"),
				"version": c.String("version"),
			}

			finder := NsFinder{Config: cfg}
			app, env, workspace, err := finder.GetAppAndWorkspace(appName, c.String("stack-name"), envName)
			if err != nil {
				return err
			}

			provider := providers.Find(workspace.Module.Category, workspace.Module.Type)
			if provider == nil {
				return fmt.Errorf("unable to push, this CLI does not support category=%s, type=%s", workspace.Module.Category, workspace.Module.Type)
			}

			return provider.Push(cfg, app, env, workspace, userConfig)
		},
	}
}
