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
			&cli.StringSliceFlag{
				Name:  "var",
				Usage: "Set variable values when issuing `up`",
			},
		},
		Action: func(c *cli.Context) error {
			varFlags := c.StringSlice("var")

			return BlockWorkspaceAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, workspace types.Workspace) error {
				if workspace.Status == types.WorkspaceStatusProvisioned {
					fmt.Println("workspace is already provisioned")
					return nil
				}

				newRunConfig, err := runs.GetPromotion(cfg, workspace, "")
				if err != nil {
					return err
				}
				skipped, err := runs.SetRunConfigVars(newRunConfig, varFlags)
				if len(skipped) > 0 {
					fmt.Printf("[Warning] The following variables were skipped because they don't exist in the module: %s\n\n", strings.Join(skipped, ", "))
				}
				if err != nil {
					return err
				}

				t := true
				newRun, err := runs.Create(cfg, workspace, newRunConfig, &t, false)
				if err != nil {
					return fmt.Errorf("error creating run: %w", err)
				} else if newRun == nil {
					return fmt.Errorf("unable to create run")
				}
				fmt.Printf("created run %q\n", newRun.Uid)
				fmt.Fprintln(os.Stdout, runs.GetBrowserUrl(cfg, workspace, *newRun))

				if c.IsSet("wait") {
					return runs.StreamLogs(ctx, cfg, workspace, newRun)
				}
				return nil
			})
		},
	}
}
