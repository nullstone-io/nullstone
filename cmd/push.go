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
				Name: "stack",
				Usage: `The stack name where the app resides.
       This is only required if multiple apps have the same 'app-name'.`,
			},
			cli.StringFlag{
				Name: "source",
				Usage: `The source artifact to push.
       app/container: This is the docker image to push. This follows the same syntax as 'docker push NAME[:TAG]'.
       app/serverless: This is a .zip archive to push.`,
				Required: true,
			},
			cli.StringFlag{
				Name: "image-tag",
				Usage: `Push the image with this tag instead of the source. 
       If not specified, will use the source tag.
       This is only used for app/container applications.`,
			},
			cli.StringFlag{
				Name: "version",
				Usage: `Push the artifact with this version.
       app/container: This is not used.
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
				"source":   c.String("source"),
				"imageTag": c.String("image-tag"),
				"version":  c.String("version"),
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
