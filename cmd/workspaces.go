package cmd

import (
	"context"
	"fmt"
	"maps"
	"os"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/modules"
	"gopkg.in/nullstone-io/nullstone.v0/tfconfig"
	"gopkg.in/nullstone-io/nullstone.v0/workspaces"
)

var Workspaces = &cli.Command{
	Name:      "workspaces",
	Usage:     "View and modify workspaces",
	UsageText: "nullstone workspaces [subcommand]",
	Subcommands: []*cli.Command{
		WorkspacesSelect,
	},
}

const (
	DefaultToolName = "terraform"
)

var WorkspacesSelect = &cli.Command{
	Name:        "select",
	Description: "Sync a given workspace's state with the current directory. Running this command will allow you to run terraform plans/applies locally against the selected workspace.",
	Usage:       "Select workspace",
	UsageText:   "nullstone workspaces select [--stack=<stack>] --block=<block> --env=<env> [--config=<effective|latest|current>]",
	Flags: []cli.Flag{
		StackFlag,
		&cli.StringFlag{
			Name:     "block",
			Usage:    "Name of the block to use for this operation",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "env",
			Usage:    `Name of the environment to use for this operation`,
			Required: true,
		},
		&cli.StringFlag{
			Name:  "config",
			Usage: "Which workspace configuration to sync locally. `effective` (default) is the latest configuration including unapplied changes queued in the Nullstone UI. `latest` is the last applied configuration; its run may still be in progress. `current` is the configuration from the last finished run.",
			Value: string(workspaces.ConfigSourceEffective),
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		configSource, err := workspaces.ParseConfigSource(c.String("config"))
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}
		return ProfileAction(c, func(cfg api.Config) error {
			toolName := detectModuleToolName()

			if !tfconfig.IsCredsConfigured(cfg) {
				if err := tfconfig.ConfigCreds(ctx, cfg); err != nil {
					fmt.Printf("Warning: unable to configure Terraform-based credentials with Nullstone servers: %s\n", err)
				} else {
					fmt.Println("Configured Terraform-based credentials with Nullstone servers.")
				}
			}

			client := api.Client{Config: cfg}
			stackName := c.String("stack")
			blockName := c.String("block")
			envName := c.String("env")
			sbe, err := find.StackBlockEnvByName(ctx, cfg, stackName, blockName, envName)
			if err != nil {
				return err
			}

			targetWorkspace := workspaces.Manifest{
				OrgName:      cfg.OrgName,
				StackId:      sbe.Stack.Id,
				StackName:    sbe.Stack.Name,
				BlockId:      sbe.Block.Id,
				BlockName:    sbe.Block.Name,
				BlockRef:     sbe.Block.Reference,
				EnvId:        sbe.Env.Id,
				EnvName:      sbe.Env.Name,
				Connections:  workspaces.ManifestConnections{},
				Capabilities: workspaces.ManifestCapabilities{},
			}
			workspace, err := client.Workspaces().Get(ctx, targetWorkspace.StackId, targetWorkspace.BlockId, targetWorkspace.EnvId)
			if err != nil {
				return err
			} else if workspace == nil {
				return fmt.Errorf("could not find workspace (stack=%s, block=%s, env=%s)", stackName, blockName, envName)
			}
			targetWorkspace.WorkspaceUid = workspace.Uid.String()

			config, err := workspaces.GetWorkspaceConfig(ctx, cfg, targetWorkspace, configSource)
			if err != nil {
				return fmt.Errorf("could not retrieve %s workspace configuration: %w", configSource, err)
			}
			targetWorkspace.ClassificationLevel = string(config.Metadata.DataClassification)
			fmt.Printf("Using %s workspace configuration\n", configSource)

			// Record every resolved app-level connection from the selected configuration,
			// then survey the user for any that still have no target
			targetWorkspace.Connections = workspaces.ManifestConnectionsFrom(config.Connections)
			manualConnections, err := surveyMissingConnections(ctx, cfg, targetWorkspace.StackName, "", config.Connections)
			if err != nil {
				return err
			}
			maps.Copy(targetWorkspace.Connections, workspaces.ManifestConnectionsFrom(manualConnections))

			// Same for each capability
			for _, cap := range config.Capabilities {
				capManualConnections, err := surveyMissingConnections(ctx, cfg, targetWorkspace.StackName, cap.Name, cap.Connections)
				if err != nil {
					return err
				}
				mc := workspaces.ManifestCapability{
					Connections: workspaces.ManifestConnectionsFrom(cap.Connections),
				}
				maps.Copy(mc.Connections, workspaces.ManifestConnectionsFrom(capManualConnections))
				targetWorkspace.Capabilities[cap.Name] = mc
			}

			return CancellableAction(func(ctx context.Context) error {
				return workspaces.Select(ctx, cfg, targetWorkspace, config, toolName)
			})
		})
	},
}

func detectModuleToolName() string {
	manifest, err := modules.ManifestFromFile(moduleManifestFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[warning]error reading module manifest (%s): %s\n", moduleManifestFilename, err)
		return DefaultToolName
	}
	if manifest.ToolName != "" {
		return manifest.ToolName
	}
	return DefaultToolName
}

func surveyMissingConnections(ctx context.Context, cfg api.Config, sourceStackName string, capabilityName string, conns types.Connections) (types.Connections, error) {
	initialPrompt := &sync.Once{}
	connections := types.Connections{}
	for name, conn := range conns {
		// Let's ask the user if the connection has no reference
		if conn.EffectiveTarget == nil || conn.EffectiveTarget.BlockId < 1 {
			initialPrompt.Do(func() {
				if capabilityName != "" {
					fmt.Printf("There are connections in capability %q that do not have a target set.\n", capabilityName)
				} else {
					fmt.Println("There are connections in this module that do not have a target set.")
				}
				fmt.Println("Type the block name for each connection to configure the connection locally.")
			})
			ct, err := surveyMissingConnection(ctx, cfg, sourceStackName, name, conn)
			if err != nil {
				return nil, err
			} else if ct != nil {
				connections[name] = types.Connection{
					Connection:      conn.Connection,
					EffectiveTarget: ct,
				}
			}
		}
	}
	return connections, nil
}

func surveyMissingConnection(ctx context.Context, cfg api.Config, sourceStackName, name string, conn types.Connection) (*types.ConnectionTarget, error) {
	preface := "[required]"
	if conn.Optional {
		preface = "[optional]"
	}
	input := &survey.Input{
		Message: fmt.Sprintf("%s connection %q (type=%s):", preface, name, conn.Type),
	}
	for {
		var answer string
		if err := survey.AskOne(input, &answer); err != nil {
			return nil, err
		}
		if answer == "" && conn.Optional {
			return nil, nil
		}

		ct, err := find.ConnectionTarget(ctx, cfg, sourceStackName, answer)
		if err != nil {
			fmt.Printf("Invalid connection: %s\n", err)
			fmt.Println("Try again.")
			continue
		}
		return ct, nil
	}
}
