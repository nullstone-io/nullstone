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
			&cli.StringFlag{
				Name:  "module-version",
				Usage: "Use a specific module version to run the plan.",
			},
		},
		Action: func(c *cli.Context) error {
			varFlags := c.StringSlice("var")
			moduleVersion := c.String("module-version")

			return BlockWorkspaceAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, workspace types.Workspace) error {
				moduleSourceOverride := ""
				if moduleVersion != "" {
					moduleSourceOverride = fmt.Sprintf("%s@%s", block.ModuleSource, moduleVersion)
				}
				newRunConfig, err := runs.GetPromotion(cfg, workspace, moduleSourceOverride)
				if err != nil {
					return fmt.Errorf("error getting run configuration for plan: %w", err)
				}

				skipped, err := runs.SetRunConfigVars(newRunConfig, varFlags)
				if len(skipped) > 0 {
					fmt.Printf("[Warning] The following variables were skipped because they don't exist in the module: %s\n\n", strings.Join(skipped, ", "))
				}
				if err != nil {
					return err
				}

				f := false
				newRun, err := runs.Create(cfg, workspace, newRunConfig, &f, false)
				if err != nil {
					return fmt.Errorf("error creating run: %w", err)
				} else if newRun == nil {
					return fmt.Errorf("unable to create run")
				}
				fmt.Fprintf(os.Stdout, "created run %q\n", newRun.Uid)

				if c.IsSet("wait") {
					return runs.StreamLogs(ctx, cfg, workspace, newRun)
				}
				return nil
			})
		},
	}
}
