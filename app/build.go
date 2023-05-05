package app

import (
	allApp "github.com/nullstone-io/deployment-sdk/app/all"
	"github.com/urfave/cli/v2"
	allAdmin "gopkg.in/nullstone-io/nullstone.v0/admin/all"
	"gopkg.in/nullstone-io/nullstone.v0/cmd"
	"sort"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func Build() *cli.App {
	appProviders := allApp.Providers
	adminProviders := allAdmin.Providers

	cliApp := cli.NewApp()
	cliApp.Version = version
	cliApp.Metadata = map[string]interface{}{
		"commit":  commit,
		"date":    date,
		"builtBy": builtBy,
	}
	cliApp.EnableBashCompletion = true
	cliApp.Flags = []cli.Flag{
		cmd.ProfileFlag,
		cmd.OrgFlag,
	}
	sort.Sort(cli.FlagsByName(cliApp.Flags))
	cliApp.Commands = []*cli.Command{
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
		cmd.Profile,
	}
	sort.Sort(cli.CommandsByName(cliApp.Commands))

	return cliApp
}
