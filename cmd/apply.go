package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
	"os"
	"strings"
)

var Apply = func() *cli.Command {
	return &cli.Command{
		Name:      "apply",
		Usage:     "Runs an apply with optional auto-approval",
		UsageText: "nullstone apply [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			BlockFlag,
			EnvFlag,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Stream the Terraform logs while waiting for Nullstone to run the apply.",
			},
			&cli.BoolFlag{
				Name:  "auto-approve",
				Usage: "Auto-approve any changes made in Terraform",
			},
			&cli.StringSliceFlag{
				Name:  "var",
				Usage: "Set variable values when issuing `apply`",
			},
			&cli.StringFlag{
				Name:  "module-version",
				Usage: "Use a specific module version to run the apply.",
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
				moduleSourceOverride := ""
				if moduleVersion != "" {
					moduleSourceOverride = fmt.Sprintf("%s@%s", block.ModuleSource, moduleVersion)
				}
				newRunConfig, err := runs.GetPromotion(cfg, workspace, moduleSourceOverride)
				if err != nil {
					return fmt.Errorf("error getting run configuration for apply: %w", err)
				}

				skipped, err := runs.SetRunConfigVars(newRunConfig, varFlags)
				if len(skipped) > 0 {
					fmt.Printf("[Warning] The following variables were skipped because they don't exist in the module: %s\n\n", strings.Join(skipped, ", "))
				}
				if err != nil {
					return err
				}

				newRun, err := runs.Create(cfg, workspace, newRunConfig, autoApprove, false)
				if err != nil {
					return fmt.Errorf("error creating run: %w", err)
				} else if newRun == nil {
					return fmt.Errorf("unable to create run")
				}
				fmt.Fprintf(os.Stdout, "created apply run %q\n", newRun.Uid)

				if c.IsSet("wait") {
					return runs.StreamLogs(ctx, cfg, workspace, newRun)
				}
				return nil
			})
		},
	}
}
