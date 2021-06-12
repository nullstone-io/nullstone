package cmd

import (
	"fmt"
	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"sort"
)

var Envs = &cli.Command{
	Name:      "envs",
	Usage:     "View and modify environments",
	UsageText: "nullstone envs [subcommand]",
	Subcommands: []*cli.Command{
		EnvsList,
	},
}

var EnvsList = &cli.Command{
	Name:      "list",
	Usage:     "List environments",
	UsageText: "nullstone envs list <stack-name>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "detail",
			Aliases: []string{"d"},
		},
	},
	Action: func(c *cli.Context) error {
		_, cfg, err := SetupProfileCmd(c)
		if err != nil {
			return err
		}

		if c.NArg() != 1 {
			cli.ShowCommandHelp(c, c.Command.Name)
			return fmt.Errorf("stack-name is required to list environments")
		}
		stackName := c.Args().Get(0)

		client := api.Client{Config: cfg}
		envs, err := client.EnvironmentsByName().List(stackName)
		if err != nil {
			return fmt.Errorf("error listing environments: %w", err)
		}
		sort.SliceStable(envs, func(i, j int) bool {
			return envs[i].PipelineOrder < envs[i].PipelineOrder
		})

		if c.IsSet("detail") {
			envDetails := make([]string, len(envs)+1)
			envDetails[0] = "ID|Name"
			for i, env := range envs {
				envDetails[i+1] = fmt.Sprintf("%d|%s", env.Id, env.Name)
			}
			fmt.Println(columnize.Format(envDetails, columnize.DefaultConfig()))
		} else {
			for _, env := range envs {
				fmt.Println(env.Name)
			}
		}

		return nil
	},
}
