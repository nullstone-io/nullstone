package cmd

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/tfconfig"
	"gopkg.in/nullstone-io/nullstone.v0/workspaces"
	"sync"
)

var Workspaces = &cli.Command{
	Name:      "workspaces",
	Usage:     "View and modify workspaces",
	UsageText: "nullstone workspaces [subcommand]",
	Subcommands: []*cli.Command{
		WorkspacesSelect,
	},
}

var WorkspacesSelect = &cli.Command{
	Name:      "select",
	Usage:     "Select workspace",
	UsageText: "nullstone workspaces select [--stack=<stack>] --block=<block> --env=<env>",
	Flags: []cli.Flag{
		StackFlag,
		&cli.StringFlag{
			Name:     "block",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "env",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			if !tfconfig.IsCredsConfigured(cfg) {
				if err := tfconfig.ConfigCreds(cfg); err != nil {
					fmt.Printf("Warning: unable to configure Terraform-based credentials with Nullstone servers: %s\n", err)
				} else {
					fmt.Println("Configured Terraform-based credentials with Nullstone servers.")
				}
			}

			client := api.Client{Config: cfg}
			stackName := c.String("stack")
			blockName := c.String("block")
			envName := c.String("env")
			sbe, err := find.StackBlockEnvByName(cfg, stackName, blockName, envName)
			if err != nil {
				return err
			}

			// TODO: Add support for capability testing --> workspaces.Manifest.CapabilityId

			targetWorkspace := workspaces.Manifest{
				OrgName:     cfg.OrgName,
				StackId:     sbe.Stack.Id,
				StackName:   sbe.Stack.Name,
				BlockId:     sbe.Block.Id,
				BlockName:   sbe.Block.Name,
				BlockRef:    sbe.Block.Reference,
				EnvId:       sbe.Env.Id,
				EnvName:     sbe.Env.Name,
				Connections: workspaces.ManifestConnections{},
			}
			workspace, err := client.Workspaces().Get(targetWorkspace.StackId, targetWorkspace.BlockId, targetWorkspace.EnvId)
			if err != nil {
				return err
			} else if workspace == nil {
				return fmt.Errorf("could not find workspace (stack=%s, block=%s, env=%s)", stackName, blockName, envName)
			}
			targetWorkspace.WorkspaceUid = workspace.Uid.String()

			runConfig, err := workspaces.GetRunConfig(cfg, targetWorkspace)
			if err != nil {
				return fmt.Errorf("could not retreive current workspace configuration: %w", err)
			}
			manualConnections, err := surveyMissingConnections(cfg, targetWorkspace.StackName, runConfig)
			if err != nil {
				return err
			}
			for name, conn := range manualConnections {
				targetWorkspace.Connections[name] = workspaces.ManifestConnectionTarget{
					StackId:   conn.Reference.StackId,
					BlockId:   conn.Reference.BlockId,
					BlockName: conn.Reference.BlockName,
					EnvId:     conn.Reference.EnvId,
				}
			}

			return CancellableAction(func(ctx context.Context) error {
				return workspaces.Select(ctx, cfg, targetWorkspace, runConfig)
			})
		})
	},
}

func surveyMissingConnections(cfg api.Config, sourceStackName string, runConfig types.RunConfig) (types.Connections, error) {
	initialPrompt := &sync.Once{}
	connections := types.Connections{}
	for name, conn := range runConfig.Connections {
		// Let's ask the user if the connection has no reference
		if conn.Reference == nil || conn.Reference.BlockId < 1 {
			initialPrompt.Do(func() {
				fmt.Println("There are connections in this module that do not have a target set.")
				fmt.Println("Type the block name for each connection to configure the connection locally.")
			})
			ct, err := surveyMissingConnection(cfg, sourceStackName, name, conn)
			if err != nil {
				return nil, err
			} else if ct != nil {
				connections[name] = types.Connection{
					Connection: conn.Connection,
					Reference:  ct,
				}
			}
		}
	}
	return connections, nil
}

func surveyMissingConnection(cfg api.Config, sourceStackName, name string, conn types.Connection) (*types.ConnectionTarget, error) {
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

		ct, err := find.ConnectionTarget(cfg, sourceStackName, answer)
		if err != nil {
			fmt.Printf("Invalid connection: %s\n", err)
			fmt.Println("Try again.")
		}
		return ct, nil
	}
}
