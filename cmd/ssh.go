package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"strings"
)

var Ssh = func(providers admin.Providers) *cli.Command {
	return &cli.Command{
		Name:        "ssh",
		Description: "SSH into a running app container or virtual machine. Use the `--forward, L` option to forward ports from remote service or hosts.",
		Usage:       "SSH into a running application.",
		UsageText:   "nullstone ssh [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvFlag,
			InstanceFlag,
			TaskFlag,
			PodFlag,
			ContainerFlag,
			&cli.StringSliceFlag{
				Name:    "forward",
				Aliases: []string{"L"},
				Usage:   "Use this to forward ports from host to local machine. Format: <local-port>:[<remote-host>]:<remote-port>",
			},
		},
		Action: func(c *cli.Context) error {
			forwards := make([]config.PortForward, 0)
			for _, arg := range c.StringSlice("forward") {
				pf, err := config.ParsePortForward(arg)
				if err != nil {
					return fmt.Errorf("invalid format for --forward/-L: %w", err)
				}
				forwards = append(forwards, pf)
			}

			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				source := outputs.ApiRetrieverSource{Config: cfg}
				remoter, err := providers.FindRemoter(ctx, logging.StandardOsWriters{}, source, appDetails)
				if err != nil {
					return err
				} else if remoter == nil {
					module := appDetails.Module
					platform := strings.TrimSuffix(fmt.Sprintf("%s:%s", module.Platform, module.Subplatform), ":")
					return fmt.Errorf("The Nullstone CLI does not currently support the ssh command for the %q application. (Module = %s/%s, App Category = app/%s, Platform = %s)",
						appDetails.App.Name, module.OrgName, module.Name, module.Subcategory, platform)
				}
				options := admin.RemoteOptions{
					Instance:     c.String("instance"),
					Task:         c.String("task"),
					Pod:          c.String("pod"),
					Container:    c.String("container"),
					PortForwards: forwards,
				}
				return remoter.Ssh(ctx, options)
			})
		},
	}
}
