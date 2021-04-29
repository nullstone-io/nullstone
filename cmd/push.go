package cmd

import (
	"fmt"
	"github.com/urfave/cli"
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
				Name:  "stack",
				Usage: "The stack name where the app resides. This is only required if multiple apps have the same 'app-name'.",
			},
			cli.StringFlag{
				Name:     "source",
				Usage:    "The source docker image to push. This follows the same syntax as `docker push NAME[:TAG]`.",
				Required: true,
			},
			cli.StringFlag{
				Name:  "image-tag",
				Usage: "Push the image with this tag instead of the source. If not specified, will use the source tag.",
			},
		},
		Action: func(c *cli.Context) error {
			_, cfg, err := SetupProfileCmd(c)
			if err != nil {
				return err
			}

			if c.NArg() != 2 {
				return cli.ShowCommandHelp(c, "push")
			}
			appName := c.Args().Get(0)
			envName := c.Args().Get(1)
			userConfig := map[string]string{
				"source":   c.String("source"),
				"imageTag": c.String("image-tag"),
			}

			finder := NsFinder{Config: cfg}
			app, workspace, err := finder.GetAppAndWorkspace(appName, c.String("stack-name"), envName)
			if err != nil {
				return err
			}

			provider := providers.Find(workspace.Module.Category, workspace.Module.Type)
			if provider == nil {
				return fmt.Errorf("unable to push, this CLI does not support category=%s, type=%s", workspace.Module.Category, workspace.Module.Type)
			}

			return provider.Push(cfg, app, workspace, userConfig)
		},
	}
}
