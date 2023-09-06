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

var Down = func() *cli.Command {
	return &cli.Command{
		Name:        "down",
		Description: "Destroys the infrastructure for the given block/environment.",
		Usage:       "Destroy the block infrastructure",
		UsageText:   "nullstone down [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			BlockFlag,
			EnvFlag,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait for the launch to complete and stream the Terraform logs to the console.",
			},
		},
		Action: func(c *cli.Context) error {
			return BlockWorkspaceAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, workspace types.Workspace) error {
				if workspace.Status == types.WorkspaceStatusNotProvisioned {
					fmt.Println("workspace is not launched")
					return nil
				}

				f := false
				newRun, err := runs.Create(cfg, workspace, &f, true)
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
