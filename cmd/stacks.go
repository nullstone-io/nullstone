package cmd

import (
	"fmt"
	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var Stacks = &cli.Command{
	Name:      "stacks",
	Usage:     "View and modify stacks",
	UsageText: "nullstone stacks [subcommand]",
	Subcommands: []*cli.Command{
		StacksList,
		StacksNew,
	},
}

var StacksList = &cli.Command{
	Name:        "list",
	Description: "Shows a list of the stacks that you have access to. Set the `--detail` flag to show more details about each stack.",
	Usage:       "List stacks",
	UsageText:   "nullstone stacks list",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "detail",
			Aliases: []string{"d"},
			Usage:   "Use this flag to show more details about each stack",
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			allStacks, err := client.Stacks().List()
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
		})
	},
}

var StacksNew = &cli.Command{
	Name:        "new",
	Description: "Creates a new stack with the given name and in the organization configured for the CLI.",
	Usage:       "Create new stack",
	UsageText:   "nullstone stacks new --name=<name> --description=<description>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Usage:    "The name of the stack to create. This name must be unique within the organization.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "description",
			Usage:    "The description of the stack to create.",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			name := c.String("name")
			description := c.String("description")
			stack, err := client.Stacks().Create(&types.Stack{
				Name:        name,
				Description: description,
			})
			if err != nil {
				return fmt.Errorf("error creating stack: %w", err)
			}
			fmt.Printf("created %q stack\n", stack.Name)
			return nil
		})
	},
}
