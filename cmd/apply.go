package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/modules"
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
				Usage: "The version of the module to apply. When used with `--publish`, this specifies the version to publish (supports semver, `next-patch`, and `next-build`).",
			},
			&cli.BoolFlag{
				Name:  "publish",
				Usage: "Package and publish the module in the current directory before running the apply. Uses `next-build` for the version by default, or the value of `--module-version` if specified.",
			},
		},
		Action: func(c *cli.Context) error {
			varFlags := c.StringSlice("var")
			moduleVersion := c.String("module-version")
			publish := c.Bool("publish")
			var autoApprove *bool
			if c.IsSet("auto-approve") {
				val := c.Bool("auto-approve")
				autoApprove = &val
			}
			logger := log.New(os.Stderr, "", 0)

			return BlockWorkspaceAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, workspace types.Workspace) error {
				// Publish phase
				if publish {
					input := modules.PublishInput{Version: moduleVersion}
					if input.Version == "" {
						input.Version = "next-build"
					}
					output, err := modules.Publish(ctx, cfg, logger, input)
					if err != nil {
						return cli.Exit(fmt.Sprintf("publish failed: %s", err), 3)
					}
					// Use the published version for the run
					moduleVersion = output.Version
				}

				if moduleVersion != "" {
					input := api.UpdateWorkspaceModuleInput{ModuleVersion: moduleVersion}
					err := runs.SetModuleVersion(ctx, cfg, workspace, input)
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
				err = PerformRun(ctx, cfg, input)

				if err == nil {
					return nil
				}

				// Map run errors to specific exit codes
				if errors.Is(err, runs.ErrRunDisapproved) {
					return nil
				}
				var runErr *runs.RunFailedError
				if errors.As(err, &runErr) {
					if runErr.Phase == "apply" {
						return cli.Exit(fmt.Sprintf("apply failed: %s", runErr.StatusMessage), 2)
					}
					return cli.Exit(fmt.Sprintf("plan failed: %s", runErr.StatusMessage), 1)
				}
				return err
			})
		},
	}
}
