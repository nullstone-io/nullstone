package cmd

import (
	"context"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
)

var Apply = func() *cli.Command {
	return &cli.Command{
		Name: "apply",
		Description: "Runs a Terraform apply on the given block and environment. This is useful for making ad-hoc changes to your infrastructure.\n" +
			"This plan will be executed by the Nullstone system. In order to run a plan locally, check out the `nullstone workspaces select` command.\n" +
			"Be sure to run `nullstone plan` first to see what changes will be made.",
		Usage:     "Runs an apply with optional auto-approval",
		UsageText: "nullstone apply [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			BlockFlag,
			EnvFlag,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait for the apply to complete and stream the Terraform logs to the console.",
			},
			&cli.BoolFlag{
				Name:  "auto-approve",
				Usage: "Skip any approvals and apply the changes immediately. This requires proper permissions in the stack.",
			},
			&cli.StringSliceFlag{
				Name:  "var",
				Usage: "Set variables values for the apply. This can be used to override variables defined in the module.",
			},
			&cli.StringFlag{
				Name:  "module-version",
				Usage: "The version of the module to apply.",
			},
		},
		Action: func(c *cli.Context) error {
			varFlags := c.StringSlice("var")
			moduleVersion := c.String("module-version")
			var autoApprove *bool
			if c.IsSet("auto-approve") {
				val := c.Bool("auto-approve")
				autoApprove = &val
			}

			return BlockWorkspaceAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, workspace types.Workspace) error {
				if moduleVersion != "" {
					module := types.WorkspaceModuleInput{
						Module:        block.ModuleSource,
						ModuleVersion: moduleVersion,
					}
					err := runs.SetModuleVersion(ctx, cfg, workspace, module)
					if err != nil {
						return err
					}
				}

				err := runs.SetConfigVars(ctx, cfg, workspace, varFlags)
				if err != nil {
					return err
				}

				input := PerformRunInput{
					Workspace:  workspace,
					CommitSha:  "",
					IsApproved: autoApprove,
					IsDestroy:  false,
					BlockType:  types.BlockType(block.Type),
					StreamLogs: c.IsSet("wait"),
				}
				return PerformRun(ctx, cfg, input)
			})
		},
	}
}
