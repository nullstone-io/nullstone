package main

import (
	"context"
	"errors"
	"fmt"
	allApp "github.com/nullstone-io/deployment-sdk/app/all"
	"github.com/urfave/cli/v2"
	allAdmin "gopkg.in/nullstone-io/nullstone.v0/admin/all"
	"gopkg.in/nullstone-io/nullstone.v0/cmd"
	"os"
	"sort"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	appProviders := allApp.Providers
	adminProviders := allAdmin.Providers

	cliApp := &cli.App{
		Version:              version,
		EnableBashCompletion: true,
		Metadata: map[string]interface{}{
			"commit":  commit,
			"date":    date,
			"builtBy": builtBy,
		},
		Flags: []cli.Flag{
			cmd.ProfileFlag,
			cmd.OrgFlag,
		},
		Commands: []*cli.Command{
			{
				Name: "version",
				Action: func(c *cli.Context) error {
					cli.ShowVersion(c)
					return nil
				},
			},
			cmd.Configure,
			cmd.SetOrg,
			cmd.Stacks,
			cmd.Envs,
			cmd.Apps,
			cmd.Blocks,
			cmd.Modules,
			cmd.Workspaces,
			cmd.Up(),
			cmd.Plan(),
			cmd.Apply(),
			cmd.Outputs(),
			cmd.Push(appProviders),
			cmd.Deploy(appProviders),
			cmd.Launch(appProviders),
			cmd.Logs(adminProviders),
			cmd.Status(adminProviders),
			cmd.Exec(adminProviders),
			cmd.Ssh(adminProviders),
		},
	}
	sort.Sort(cli.FlagsByName(cliApp.Flags))
	sort.Sort(cli.CommandsByName(cliApp.Commands))

	err := cliApp.Run(os.Args)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
