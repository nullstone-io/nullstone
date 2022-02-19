package cmd

import (
	"fmt"
	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var Blocks = &cli.Command{
	Name:      "blocks",
	Usage:     "View and modify blocks",
	UsageText: "nullstone blocks [subcommand]",
	Subcommands: []*cli.Command{
		BlocksList,
	},
}

var BlocksList = &cli.Command{
	Name:      "list",
	Usage:     "List blocks",
	UsageText: "nullstone blocks list --stack=<stack>",
	Flags: []cli.Flag{
		StackRequiredFlag,
		&cli.BoolFlag{
			Name:    "detail",
			Aliases: []string{"d"},
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}

			stackName := c.String("stack")
			stack, err := client.StacksByName().Get(stackName)
			if err != nil {
				return fmt.Errorf("error looking for stack %q: %w", stackName, err)
			} else if stack == nil {
				return fmt.Errorf("stack %q does not exist in organization %q", stackName, cfg.OrgName)
			}

			allBlocks, err := client.Blocks().List(stack.Id)
			if err != nil {
				return fmt.Errorf("error listing blocks: %w", err)
			}

			if c.IsSet("detail") {
				appDetails := make([]string, len(allBlocks)+1)
				appDetails[0] = "ID|Type|Name|Reference|Category|Module Type|Module|Stack"
				for i, block := range allBlocks {
					var blockCategory types.CategoryName
					var blockType string
					if blockModule, err := find.BlockModule(cfg, block); err == nil {
						blockCategory = blockModule.Category
						blockType = blockModule.Type
					}
					appDetails[i+1] = fmt.Sprintf("%d|%s|%s|%s|%s|%s|%s|%s", block.Id, block.Type, block.Name, block.Reference, blockCategory, blockType, block.ModuleSource, block.StackName)
				}
				fmt.Println(columnize.Format(appDetails, columnize.DefaultConfig()))
			} else {
				for _, block := range allBlocks {
					fmt.Println(block.Name)
				}
			}

			return nil
		})
	},
}
