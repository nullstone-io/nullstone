package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
	"os"
)

var Plan = func() *cli.Command {
	return &cli.Command{
		Name:        "plan",
		Description: "Run a plan for a given block and environment. This will automatically disapprove the plan and is useful for testing what a plan will do.",
		Usage:       "Runs a plan with a disapproval",
		UsageText:   "nullstone plan [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			BlockFlag,
			EnvFlag,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait for the plan to complete and stream the Terraform logs to the console.",
			},
			&cli.StringSliceFlag{
				Name:  "var",
				Usage: "Set variables values for the plan. This can be used to override variables defined in the module.",
			},
			&cli.StringFlag{
				Name:  "module-version",
				Usage: "Run a plan with a specific version of the module.",
			},
		},
		Action: func(c *cli.Context) error {
			varFlags := c.StringSlice("var")
			moduleVersion := c.String("module-version")

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

				f := false
				newRun, err := runs.Create(ctx, cfg, workspace, &f, false)
				if err != nil {
					return fmt.Errorf("error creating run: %w", err)
				} else if newRun == nil {
					return fmt.Errorf("unable to create run")
				}
				fmt.Fprintf(os.Stdout, "created plan run %q\n", newRun.Uid)
				fmt.Fprintln(os.Stdout, runs.GetBrowserUrl(cfg, workspace, *newRun))

				if c.IsSet("wait") {
					err := runs.StreamLogs(ctx, cfg, workspace, newRun)
					if err == runs.ErrRunDisapproved {
						// Disapproved plans are expected, return no error
						return nil
					}
					return err
				}
				return nil
			})
		},
	}
}
