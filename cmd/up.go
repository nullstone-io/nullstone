package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var Up = func() *cli.Command {
	return &cli.Command{
		Name:      "up",
		Usage:     "Provisions the block and all of its dependencies",
		UsageText: "nullstone up [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			BlockFlag,
			EnvFlag,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait for Nullstone to fully provision the workspace.",
			},
		},
		Action: func(c *cli.Context) error {
			return BlockEnvAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment) error {
				client := api.Client{Config: cfg}
				workspace, err := client.Workspaces().Get(stack.Id, block.Id, env.Id)
				if err != nil {
					return fmt.Errorf("error looking for workspace: %w", err)
				} else if workspace == nil {
					return fmt.Errorf("workspace not found")
				}

				if workspace.Status == types.WorkspaceStatusProvisioned {
					fmt.Println("workspace is already provisioned")
					return nil
				}

				newRunConfig, err := client.PromotionConfigs().Get(workspace.StackId, workspace.BlockId, workspace.EnvId)
				if err != nil {
					return err
				}

				isApproved := true
				input := types.CreateRunInput{
					IsDestroy:     false,
					IsApproved:    &isApproved,
					Source:        newRunConfig.Source,
					SourceVersion: newRunConfig.SourceVersion,
					Variables:     newRunConfig.Variables,
					EnvVariables:  newRunConfig.EnvVariables,
					Connections:   newRunConfig.Connections,
					Capabilities:  newRunConfig.Capabilities,
					Providers:     newRunConfig.Providers,
				}

				newRun, err := client.Runs().Create(workspace.StackId, workspace.Uid, input)
				if err != nil {
					return fmt.Errorf("error creating run: %w", err)
				} else if newRun == nil {
					return fmt.Errorf("unable to create run")
				}
				fmt.Println("run created")
				return nil
			})
		},
	}
}
