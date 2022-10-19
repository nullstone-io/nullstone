package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"strings"
)

var Ssh = func(providers admin.Providers) *cli.Command {
	return &cli.Command{
		Name:      "ssh",
		Usage:     "SSH into a running service. Use to forward ports from remote service or hosts.",
		UsageText: "nullstone ssh [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvFlag,
			TaskFlag,
			&cli.StringSliceFlag{
				Name:    "forward",
				Aliases: []string{"L"},
				Usage:   "Use this to forward ports from host to local machine. Format: <local-port>:[<remote-host>]:<remote-port>",
			},
		},
		Action: func(c *cli.Context) error {
			task := c.String("task")

			forwards := make([]config.PortForward, 0)
			for _, arg := range c.StringSlice("forward") {
				pf, err := config.ParsePortForward(arg)
				if err != nil {
					return fmt.Errorf("invalid format for --forward/-L: %w", err)
				}
				forwards = append(forwards, pf)
			}

			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				remoter, err := providers.FindRemoter(logging.StandardOsWriters{}, cfg, appDetails)
				if err != nil {
					return err
				} else if remoter == nil {
					module := appDetails.Module
					platform := strings.TrimSuffix(fmt.Sprintf("%s:%s", module.Platform, module.Subplatform), ":")
					return fmt.Errorf("The Nullstone CLI does not currently support the ssh command for the %q application. (Module = %s, App Category = app/%s, Platform = %s)",
						appDetails.App.Name, module.OrgName, module.Name, module.Subcategory, platform)
				}
				return remoter.Ssh(ctx, task, forwards)
			})
		},
	}
}
