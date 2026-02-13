package app

import (
	allApp "github.com/nullstone-io/deployment-sdk/app/all"
	"github.com/urfave/cli/v2"
	allAdmin "gopkg.in/nullstone-io/nullstone.v0/admin/all"
	"gopkg.in/nullstone-io/nullstone.v0/cmd"
	"sort"
)

func Build() *cli.App {
	appProviders := allApp.Providers
	adminProviders := allAdmin.Providers

	cliApp := cli.NewApp()
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
		cmd.Iac,
		cmd.Up(),
		cmd.Plan(),
		cmd.Apply(),
		cmd.Outputs(),
		cmd.Wait(),
		cmd.Push(appProviders),
		cmd.Deploy(appProviders),
		cmd.Launch(appProviders),
		cmd.Logs(appProviders),
		cmd.Status(adminProviders),
		cmd.Exec(appProviders, adminProviders),
		cmd.Ssh(adminProviders),
		cmd.Run(appProviders, adminProviders),
		cmd.Profile,
		cmd.McpServer,
	}
	sort.Sort(cli.CommandsByName(cliApp.Commands))

	return cliApp
}
