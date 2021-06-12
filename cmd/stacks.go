package cmd

import (
	"fmt"
	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

var Stacks = &cli.Command{
	Name:      "stacks",
	Usage:     "View and modify stacks",
	UsageText: "nullstone stacks [subcommand]",
	Subcommands: []*cli.Command{
		StacksList,
	},
}

var StacksList = &cli.Command{
	Name:      "list",
	Usage:     "List stacks",
	UsageText: "nullstone stacks list",
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

		client := api.Client{Config: cfg}
		allStacks, err := client.StacksByName().List()
		if err != nil {
			return fmt.Errorf("error listing stacks: %w", err)
		}

		if c.IsSet("detail") {
			stackDetails := make([]string, len(allStacks)+1)
			stackDetails[0] = "ID|Name|Description"
			for i, stack := range allStacks {
				stackDetails[i+1] = fmt.Sprintf("%d|%s|%s", stack.Id, stack.Name, stack.Description)
			}
			fmt.Println(columnize.Format(stackDetails, columnize.DefaultConfig()))
		} else {
			for _, stack := range allStacks {
				fmt.Println(stack.Name)
			}
		}

		return nil
	},
}
