package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	api_runs "gopkg.in/nullstone-io/go-api-client.v0/runs"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app_urls"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
	"os"
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
				newRun, err := api_runs.Create(ctx, cfg, workspace, "", &t, false, "")
				if err != nil {
					return fmt.Errorf("error creating run: %w", err)
				} else if newRun == nil {
					return fmt.Errorf("unable to create run")
				}
				fmt.Printf("created run %q\n", newRun.Uid)
				fmt.Fprintln(os.Stdout, app_urls.GetRun(cfg, workspace, *newRun)

				if c.IsSet("wait") {
					return runs.StreamLogs(ctx, cfg, workspace, newRun)
				}
				return nil
			})
		},
	}
}
