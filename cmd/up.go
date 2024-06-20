package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
)

var Up = func() *cli.Command {
	return &cli.Command{
		Name:        "up",
		Description: "Launches the infrastructure for the given block/environment and its dependencies.",
		Usage:       "Provisions the block and all of its dependencies",
		UsageText:   "nullstone up [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			BlockFlag,
			EnvFlag,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait for the launch to complete and stream the Terraform logs to the console.",
			},
			&cli.StringSliceFlag{
				Name:  "var",
				Usage: "Set variables values for the plan. This can be used to override variables defined in the module.",
			},
		},
		Action: func(c *cli.Context) error {
			varFlags := c.StringSlice("var")

			return BlockWorkspaceAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, workspace types.Workspace) error {
				if workspace.Status == types.WorkspaceStatusProvisioned {
					fmt.Println("workspace is already provisioned")
					return nil
				}

				err := runs.SetConfigVars(ctx, cfg, workspace, varFlags)
				if err != nil {
					return err
				}

				t := true
				input := PerformRunInput{
					Workspace:  workspace,
					CommitSha:  "",
					IsApproved: &t,
					IsDestroy:  false,
					BlockType:  types.BlockType(block.Type),
					StreamLogs: c.IsSet("wait"),
				}
				return PerformRun(ctx, cfg, input)
			})
		},
	}
}
