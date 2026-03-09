package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

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
		BlocksNew,
	},
}

var BlocksList = &cli.Command{
	Name:        "list",
	Description: "Shows a list of the blocks for the given stack. Set the `--detail` flag to show more details about each block.",
	Usage:       "List blocks",
	UsageText:   "nullstone blocks list --stack=<stack>",
	Flags: []cli.Flag{
		StackRequiredFlag,
		&cli.BoolFlag{
			Name:    "detail",
			Aliases: []string{"d"},
			Usage:   "Use this flag to show more details about each block",
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			ctx := context.TODO()
			client := api.Client{Config: cfg}

			stackName := c.String(StackRequiredFlag.Name)
			stack, err := client.StacksByName().Get(ctx, stackName)
			if err != nil {
				return fmt.Errorf("error looking for stack %q: %w", stackName, err)
			} else if stack == nil {
				return fmt.Errorf("stack %q does not exist in organization %q", stackName, cfg.OrgName)
			}

			allBlocks, err := client.Blocks().List(ctx, stack.Id, false)
			if err != nil {
				return fmt.Errorf("error listing blocks: %w", err)
			}

			if c.IsSet("detail") {
				appDetails := make([]string, len(allBlocks)+1)
				appDetails[0] = "ID|Type|Name|Reference|Stack"
				for i, block := range allBlocks {
					appDetails[i+1] = fmt.Sprintf("%d|%s|%s|%s|%s", block.Id, block.Type, block.Name, block.Reference, stackName)
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
	Name:        "new",
	Description: "Creates a new block with the given name and module. If the module has any connections, you can specify them using the `--connection` parameter.",
	Usage:       "Create block",
	UsageText:   "nullstone blocks new --name=<name> --stack=<stack> --module=<module> [--connection=<connection>...]",
	Flags: []cli.Flag{
		StackRequiredFlag,
		&cli.StringFlag{
			Name:     "name",
			Required: true,
			Usage:    "Provide a name for this new block",
		},
		&cli.StringFlag{
			Name:     "module",
			Usage:    `Specify the unique name of the module to use for this block. Example: nullstone/aws-network`,
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "connection",
			Usage: "Specify any connections that this block will have to other blocks. Use the connection name as the key, and the connected block name as the value. Example: --connection network=network0",
		},
		&cli.StringFlag{
			Name: "dns-template",
			Usage: `Specify a template for the dns name portion of the subdomain.
This is a template that allows you to add "{{ NULLSTONE_ENV }}" and "{{ NULLSTONE_ORG }}" in template.
In production, the "{{ NULLSTONE_ENV }}" will be omitted to create a vanity subdomain.

Nullstone will interpolate the template to create a subdomain: "<dns-name>.<domain-name>".

If you want to create a ".nullstone.app" subdomain using an "autogen" or "nullstone-subdomain" module, set this to "{{ random() }}".

For a subdomain on your custom domain, set this to something like "api.{{ NULLSTONE_ENV }}".
- "dev" env => api.dev.example.com
- "prod" env => api.example.com
`,
			Required: false,
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			ctx := context.TODO()
			client := api.Client{Config: cfg}

			stackName := c.String(StackRequiredFlag.Name)
			stack, err := client.StacksByName().Get(ctx, stackName)
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
			module, err := find.Module(ctx, cfg, moduleSource)
			if err != nil {
				return err
			}

			connections, err := mapConnectionsToTargets(cfg, stack, connectionSlice)
			if err != nil {
				return err
			}
			if err := validateConnections(module.LatestVersion, connections); err != nil {
				return err
			}

			blockType, err := blockTypeFromModuleCategory(module.Category)
			if err != nil {
				return err
			}

			block := &types.Block{
				OrgName: cfg.OrgName,
				StackId: stack.Id,
				Type:    string(blockType),
				Name:    name,
			}
			input := api.CreateBlockInput{
				Block: *block,
				Template: &types.WorkspaceTemplateConfig{
					Module:           moduleSource,
					ModuleConstraint: "latest",
					Connections:      connections,
				},
			}
			switch types.BlockType(block.Type) {
			case types.BlockTypeApplication:
				input.Repo = ""
				input.Framework = "other"
			case types.BlockTypeSubdomain:
				dnsTemplate := c.String("dns-template")
				if dnsTemplate == "" {
					if isNullstoneSubdomainModule(module) {
						dnsTemplate = "{{ random() }}"
						fmt.Fprintf(os.Stderr, "--dns-template was not specified; defaulting to \"{{ random() }}\" for a nullstone.app subdomain")
					} else {
						return fmt.Errorf("--dns-template is required when creating a Subdomain block")
					}
				}
				input.Template.SubdomainNameTemplate = dnsTemplate
			}

			if newBlock, err := client.Blocks().Create(ctx, stack.Id, input); err != nil {
				return err
			} else if newBlock != nil {
				fmt.Printf("created %q block\n", newBlock.Name)
			} else {
				fmt.Println("unable to create block")
			}
			return nil
		})
	},
}

func blockTypeFromModuleCategory(categoryName types.CategoryName) (types.BlockType, error) {
	switch categoryName {
	case types.CategoryApp:
		return types.BlockTypeApplication, nil
	case types.CategoryCapability:
		return types.BlockTypeBlock, fmt.Errorf("A capability module cannot be created as a standalone block. It must be attached as a capability to an application.")
	case types.CategoryDatastore:
		return types.BlockTypeDatastore, nil
	case types.CategoryIngress:
		return types.BlockTypeIngress, nil
	case types.CategorySubdomain:
		return types.BlockTypeSubdomain, nil
	case types.CategoryDomain:
		return types.BlockTypeDomain, nil
	case types.CategoryCluster:
		return types.BlockTypeCluster, nil
	case types.CategoryClusterNamespace:
		return types.BlockTypeClusterNamespace, nil
	case types.CategoryNetwork:
		return types.BlockTypeNetwork, nil
	}
	return types.BlockTypeBlock, nil
}

func mapConnectionsToTargets(cfg api.Config, stack *types.Stack, mappings []string) (map[string]types.ConnectionTarget, error) {
	ctx := context.TODO()

	connections := map[string]types.ConnectionTarget{}
	for _, connMapping := range mappings {
		tokens := strings.SplitN(connMapping, "=", 2)
		if len(tokens) < 2 {
			return nil, fmt.Errorf("invalid connection mapping %q: must specify <connection-name>=<block-name>", connMapping)
		}
		ct, err := find.ConnectionTarget(ctx, cfg, stack.Name, tokens[1])
		if err != nil {
			return nil, fmt.Errorf("error finding %q: %w", tokens[1], err)
		}
		connections[tokens[0]] = *ct
	}
	return connections, nil
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

func isNullstoneSubdomainModule(module *types.Module) bool {
	latest := module.LatestVersion
	if latest == nil {
		return false
	}
	_, hasDomain := latest.Manifest.Connections["domain"]
	_, hasSubdomain := latest.Manifest.Connections["subdomain"]
	return !hasDomain && !hasSubdomain
}
