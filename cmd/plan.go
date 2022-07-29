package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
	"strings"
)

var Plan = func() *cli.Command {
	return &cli.Command{
		Name:      "plan",
		Usage:     "Runs a plan with a disapproval",
		UsageText: "nullstone plan [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			BlockFlag,
			EnvFlag,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Stream the Terraform logs while waiting for Nullstone to run the plan.",
			},
			&cli.StringSliceFlag{
				Name:  "var",
				Usage: "Set variable values when issuing `plan`",
			},
		},
		Action: func(c *cli.Context) error {
			return BlockWorkspaceAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, workspace types.Workspace) error {
				varFlags := c.StringSlice("var")



				newRunConfig, err := client.PromotionConfigs().Get(workspace.StackId, workspace.BlockId, workspace.EnvId)
				if err != nil {
					return err
				}

				fillRunConfigVariables(newRunConfig)

				skipped, err := setRunConfigVars(newRunConfig, varFlags)
				if len(skipped) > 0 {
					fmt.Printf("[Warning] The following variables were skipped because they don't exist in the module: %s\n\n", strings.Join(skipped, ", "))
				}
				if err != nil {
					return err
				}

				isApproved := false
				input := types.CreateRunInput{
					IsDestroy:         false,
					IsApproved:        &isApproved,
					Source:            newRunConfig.Source,
					SourceVersion:     newRunConfig.SourceVersion,
					Variables:         newRunConfig.Variables,
					EnvVariables:      newRunConfig.EnvVariables,
					Connections:       newRunConfig.Connections,
					Capabilities:      newRunConfig.Capabilities,
					Providers:         newRunConfig.Providers,
					DependencyConfigs: newRunConfig.DependencyConfigs,
				}

				newRun, err := client.Runs().Create(workspace.StackId, workspace.Uid, input)
				if err != nil {
					return fmt.Errorf("error creating run: %w", err)
				} else if newRun == nil {
					return fmt.Errorf("unable to create run")
				}
				fmt.Printf("created run %q\n", newRun.Uid)

				if c.IsSet("wait") {
					return runs.StreamLogs(ctx, cfg, workspace, newRun)
				}

				return nil
			})
		},
	}
}
