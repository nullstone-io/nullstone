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
			if err := surveyMissingConnections(cfg, targetWorkspace.StackName, &runConfig); err != nil {
				return err
			}
			for name, conn := range runConfig.Connections {
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

func surveyMissingConnections(cfg api.Config, sourceStackName string, runConfig *types.RunConfig) error {
	for name, conn := range runConfig.Connections {
		// If a connection is required and does not have a target
		//   let's require the user to set a target manually
		if !conn.Optional && (conn.Reference == nil || conn.Reference.BlockId < 1) {
			input := &survey.Input{
				Message: "Connection %q is required. Choose a block to use for the connection:",
			}
			var answer string
			if err := survey.AskOne(input, &answer); err != nil {
				return err
			}
			ct, err := find.ConnectionTarget(cfg, sourceStackName, answer)
			if err != nil {
				return err
			}
			conn.Reference = ct
			// conn is byval, set it back on the map
			runConfig.Connections[name] = conn
		}
	}
	return nil
}
