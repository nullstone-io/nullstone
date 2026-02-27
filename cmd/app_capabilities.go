package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/nullstone-io/iac/yaml"
	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var AppsCapabilities = &cli.Command{
	Name:      "capabilities",
	Usage:     "View and modify app capabilities",
	UsageText: "nullstone apps capabilities [subcommand]",
	Subcommands: []*cli.Command{
		AppsCapabilitiesList,
		AppsCapabilitiesCreate,
		AppsCapabilitiesRemove,
	},
}

var AppsCapabilitiesList = &cli.Command{
	Name:        "list",
	Description: "Shows a list of capabilities for the given app. If --env is specified, lists capabilities for the app workspace. Otherwise, lists capabilities from the app's workspace template.",
	Usage:       "List app capabilities",
	UsageText:   "nullstone apps capabilities list --stack=<stack> --app=<app> [--env=<env>]",
	Flags: []cli.Flag{
		StackRequiredFlag,
		RequiredAppFlag,
		EnvOptionalFlag,
		&cli.BoolFlag{
			Name:    "detail",
			Aliases: []string{"d"},
			Usage:   "Use this flag to show more details about each capability",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			detail := c.Bool("detail")

			app, env, err := AppOrAppWorkspace(c, cfg)
			if err != nil {
				return err
			}

			if env != nil {
				// List capabilities from environment workspace
				caps, err := client.AppCapabilities().List(ctx, app.StackId, app.Id, env.Id)
				if err != nil {
					return fmt.Errorf("error listing capabilities: %w", err)
				}

				if len(caps) == 0 {
					fmt.Println("This app has no capabilities")
					return nil
				}

				if detail {
					rows := make([]string, len(caps)+1)
					rows[0] = "Name|Module|Module Version|Connections"
					for i, cur := range caps {
						rows[i+1] = fmt.Sprintf("%s|%s|%s|%s", cur.Name, cur.Source, cur.SourceVersion, formatCapabilityConnections(cur.Connections))
					}
					fmt.Println(columnize.Format(rows, columnize.DefaultConfig()))
				} else {
					for _, cur := range caps {
						fmt.Println(cur.Name)
					}
				}
			} else {
				// List capabilities from the workspace template
				wt, err := client.WorkspaceTemplates().Get(ctx, app.StackId, app.Id)
				if err != nil {
					return fmt.Errorf("error retrieving workspace template for app %q: %w", app.Name, err)
				} else if wt == nil || len(wt.Config.Capabilities) == 0 {
					fmt.Println("This app has no capabilities")
					return nil
				}

				if detail {
					rows := make([]string, len(wt.Config.Capabilities)+1)
					rows[0] = "Name|Module|Module Constraint|Connections"
					for i, cur := range wt.Config.Capabilities {
						rows[i+1] = fmt.Sprintf("%s|%s|%s|%s", cur.Name, cur.Module, cur.ModuleConstraint, formatTemplateConnections(cur.Connections))
					}
					fmt.Println(columnize.Format(rows, columnize.DefaultConfig()))
				} else {
					for _, cur := range wt.Config.Capabilities {
						fmt.Println(cur.Name)
					}
				}
			}

			return nil
		})
	},
}

var AppsCapabilitiesCreate = &cli.Command{
	Name:        "create",
	Description: "Adds a capability to an app. If --env is specified, adds the capability to that environment workspace. Otherwise, adds the capability to the app's workspace template.",
	Usage:       "Add a capability to an app",
	UsageText:   "nullstone apps capabilities create --stack=<stack> --app=<app> [--env=<env>] --module=<module> [--connection=<connection>...]",
	Flags: []cli.Flag{
		StackRequiredFlag,
		RequiredAppFlag,
		EnvOptionalFlag,
		&cli.StringFlag{
			Name:     "name",
			Usage:    `Specify the name of the capability to create.`,
			Required: true,
		},
		&cli.StringFlag{
			Name:     "module",
			Usage:    `Specify the unique name of the module to use for this capability. Example: nullstone/aws-s3-access`,
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "connection",
			Usage: "Specify any connections that this capability will have to other blocks. Use the connection name as the key, and the connected block name as the value. Example: --connection network=network0",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			name := c.String("name")
			moduleSource := c.String("module")
			connectionSlice := c.StringSlice("connection")

			app, env, err := AppOrAppWorkspace(c, cfg)
			if err != nil {
				return err
			}

			connections := types.ConnectionTargets{}
			for _, cur := range connectionSlice {
				connName, raw, ok := strings.Cut(cur, "=")
				if !ok {
					return fmt.Errorf("invalid --connection format, expected <name>=<value>")
				}
				constraint := yaml.ParseConnectionConstraint(raw)
				connections[connName] = types.ConnectionTarget{
					StackName: constraint.StackName,
					BlockName: constraint.BlockName,
					EnvName:   constraint.EnvName,
				}
			}

			if env == nil {
				// Add capability to workspace template
				capInput := types.WorkspaceCapabilityTemplateConfig{
					Name:             name,
					Module:           moduleSource,
					ModuleConstraint: "latest",
					Connections:      connections,
				}
				input := api.CreateTemplateCapabilityInput{
					Capability: capInput,
				}
				ok, err := client.WorkspaceTemplates().CreateCapability(ctx, app.StackId, app.Id, input)
				if err != nil {
					return fmt.Errorf("error creating capability in workspace template: %w", err)
				} else if !ok {
					return fmt.Errorf("unable to create capability in workspace template for app %q", app.Name)
				}
				fmt.Printf("created capability using module %q in app %q workspace template\n", moduleSource, app.Name)
			} else {
				// Add capability to environment workspace
				capInput := api.CreateCapabilityInput{
					Name:                name,
					ModuleSource:        moduleSource,
					ModuleSourceVersion: "latest",
					Connections:         connections,
				}
				caps, _, err := client.AppCapabilities().Create(ctx, app.StackId, app.Id, env.Id, []api.CreateCapabilityInput{capInput}, nil, nil)
				if err != nil {
					return fmt.Errorf("error creating capability: %w", err)
				}
				if len(caps) > 0 {
					fmt.Printf("created capability %q using module %q in app %q/%q\n", caps[0].Name, moduleSource, app.Name, env.Name)
				} else {
					fmt.Printf("created capability using module %q in app workspace %q/%q\n", moduleSource, app.Name, env.Name)
				}
			}

			return nil
		})
	},
}

var AppsCapabilitiesRemove = &cli.Command{
	Name:        "remove",
	Description: "Removes a capability from an app. If --env is specified, removes the capability from that environment workspace. Otherwise, removes the capability from the app's workspace template.",
	Usage:       "Remove a capability from an app",
	UsageText:   "nullstone apps capabilities remove --stack=<stack> --app=<app> [--env=<env>] --name=<capability-name>",
	Flags: []cli.Flag{
		StackRequiredFlag,
		RequiredAppFlag,
		EnvOptionalFlag,
		&cli.StringFlag{
			Name:     "name",
			Usage:    "The name of the capability to remove",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			capName := c.String("name")

			app, env, err := AppOrAppWorkspace(c, cfg)
			if err != nil {
				return err
			}

			if env == nil {
				// Remove capability from workspace template
				ok, err := client.WorkspaceTemplates().RemoveCapability(ctx, app.StackId, app.Id, capName)
				if err != nil {
					return fmt.Errorf("error removing capability %q from workspace template: %w", capName, err)
				} else if !ok {
					return fmt.Errorf("capability %q does not exist in app %q workspace template", capName, app.Name)
				}
				fmt.Printf("removed capability %q from app %q workspace template\n", capName, app.Name)
			} else {
				// Remove capability from environment workspace
				ok, err := client.AppCapabilities().Destroy(ctx, app.StackId, app.Id, env.Id, capName)
				if err != nil {
					return fmt.Errorf("error removing capability %q: %w", capName, err)
				} else if !ok {
					return fmt.Errorf("capability %q does not exist in app %q/%q", capName, app.Name, env.Name)
				}
				fmt.Printf("removed capability %q from app workspace %q/%q\n", capName, app.Name, env.Name)
			}

			return nil
		})
	},
}

// formatCapabilityConnections renders a Connections map (from a CapabilityConfig) as
// a comma-separated list of "name->target" pairs, e.g. "network->my-network, cluster->my-cluster".
func formatCapabilityConnections(connections types.Connections) string {
	if len(connections) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(connections))
	for name, conn := range connections {
		target := conn.EffectiveTarget
		if target == nil {
			target = conn.DesiredTarget
		}
		pairs = append(pairs, fmt.Sprintf("%s->%s", name, formatConnectionTarget(target)))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ", ")
}

// formatTemplateConnections renders a ConnectionTargets map (from a WorkspaceCapabilityTemplateConfig)
// as a comma-separated list of "name->target" pairs.
func formatTemplateConnections(connections types.ConnectionTargets) string {
	if len(connections) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(connections))
	for name, target := range connections {
		t := target
		pairs = append(pairs, fmt.Sprintf("%s->%s", name, formatConnectionTarget(&t)))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ", ")
}

// formatConnectionTarget returns the most succinct human-readable form of a ConnectionTarget:
//
//	stack.env.block  (cross-stack, cross-env)
//	stack.block      (cross-stack, same-env)
//	block            (same-stack, same-env)
func formatConnectionTarget(t *types.ConnectionTarget) string {
	if t == nil {
		return "(none)"
	}
	block := t.BlockName
	if block == "" {
		return "(none)"
	}
	if t.StackName != "" && t.EnvName != "" {
		return fmt.Sprintf("%s.%s.%s", t.StackName, t.EnvName, block)
	}
	if t.StackName != "" {
		return fmt.Sprintf("%s.%s", t.StackName, block)
	}
	return block
}
