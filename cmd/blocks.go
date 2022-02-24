package cmd

import (
	"fmt"
	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"sort"
	"strings"
)

var Blocks = &cli.Command{
	Name:      "blocks",
	Usage:     "View and modify blocks",
	UsageText: "nullstone blocks [subcommand]",
	Subcommands: []*cli.Command{
		BlocksList,
		BlocksNew,
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

			stackName := c.String(StackRequiredFlag.Name)
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
					if blockModule, err := find.Module(cfg, block.ModuleSource); err == nil {
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

var BlocksNew = &cli.Command{
	Name:      "new",
	Usage:     "Create block",
	UsageText: "nullstone blocks new --name=<name> --stack=<stack> --module=<module> [--connection=<connection>...]",
	Flags: []cli.Flag{
		StackRequiredFlag,
		&cli.StringFlag{
			Name:     "name",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "module",
			Usage:    `Specify the unique name of the module to use for this block. Example: nullstone/aws-network`,
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "connection",
			Usage: "Map the connection name on the module to the block name in the stack. Example: --connection network=network0",
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}

			stackName := c.String(StackRequiredFlag.Name)
			stack, err := client.StacksByName().Get(stackName)
			if err != nil {
				return fmt.Errorf("error looking for stack %q: %w", stackName, err)
			} else if stack == nil {
				return fmt.Errorf("stack %q does not exist in organization %q", stackName, cfg.OrgName)
			}

			name := c.String("name")
			moduleSource := c.String("module")
			if !strings.Contains(moduleSource, "/") {
				// Add organization to module source if it does not have one
				moduleSource = fmt.Sprintf("%s/%s", cfg.OrgName, moduleSource)
			}
			connectionSlice := c.StringSlice("connection")

			// TODO: Add support for module version in --module
			module, err := find.Module(cfg, moduleSource)
			if err != nil {
				return err
			}
			sort.Sort(sort.Reverse(module.Versions)) // "latest" will be at the beginning now
			var latestModuleVersion *types.ModuleVersion
			if len(module.Versions) > 0 {
				latestModuleVersion = &module.Versions[0]
			}

			connections, parentBlocks, err := mapConnectionsToTargets(cfg, stack, connectionSlice)
			if err != nil {
				return err
			}
			if err := validateConnections(latestModuleVersion, connections); err != nil {
				return err
			}

			block := &types.Block{
				Type:                blockTypeFromModuleCategory(module.Category),
				Name:                name,
				ModuleSource:        moduleSource,
				ModuleSourceVersion: "latest",
				Connections:         connections,
				ParentBlocks:        parentBlocks,
			}
			if strings.HasPrefix(string(module.Category), "app/") {
				app := &types.Application{
					Block:     *block,
					Repo:      "",
					Framework: "other",
				}
				if newApp, err := client.Apps().Create(app); err != nil {
					return err
				} else if newApp != nil {
					fmt.Printf("created %s app\n", newApp.Name)
				} else {
					fmt.Println("unable to create app")
				}
			} else {
				if newBlock, err := client.Blocks().Create(stack.Id, block); err != nil {
					return err
				} else if newBlock != nil {
					fmt.Printf("created %q block\n", newBlock.Name)
				} else {
					fmt.Println("unable to create block")
				}
			}
			return nil
		})
	},
}

func blockTypeFromModuleCategory(categoryName types.CategoryName) string {
	category := string(categoryName)
	if strings.HasPrefix(category, "app/") {
		return "Application"
	}
	if strings.HasPrefix(category, "capability/") {
		return "Block"
	}
	return strings.Title(category)
}

func mapConnectionsToTargets(cfg api.Config, stack *types.Stack, mappings []string) (map[string]types.ConnectionTarget, map[string]string, error) {
	connections := map[string]types.ConnectionTarget{}
	parentBlocks := map[string]string{}
	for _, connMapping := range mappings {
		tokens := strings.SplitN(connMapping, "=", 2)
		if len(tokens) < 2 {
			return nil, nil, fmt.Errorf("invalid connection mapping %q: must specify <connection-name>=<block-name>", connMapping)
		}
		ct, err := find.ConnectionTarget(cfg, stack.Name, tokens[1])
		if err != nil {
			return nil, nil, fmt.Errorf("error finding %q: %w", tokens[1], err)
		}
		connections[tokens[0]] = *ct
		parentBlocks[tokens[0]] = tokens[1]
	}
	return connections, parentBlocks, nil
}

func validateConnections(moduleVersion *types.ModuleVersion, connections map[string]types.ConnectionTarget) error {
	if moduleVersion == nil {
		return nil
	}

	missing := make([]string, 0)
	for name, connection := range moduleVersion.Manifest.Connections {
		if !connection.Optional {
			if _, ok := connections[name]; !ok {
				missing = append(missing, fmt.Sprintf("%s=%s", name, connection.Type))
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required connections (%s), specify using --connection", strings.Join(missing, ", "))
	}

	return nil
}
