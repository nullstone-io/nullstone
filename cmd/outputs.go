package cmd

import (
	"context"
	"encoding/json"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"os"
)

// Outputs command retrieves outputs from a workspace (block+env)
var Outputs = func() *cli.Command {
	return &cli.Command{
		Name:      "outputs",
		Usage:     "Retrieve outputs",
		UsageText: "nullstone outputs [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			BlockFlag,
			EnvFlag,
			&cli.BoolFlag{
				Name: "sensitive",
			},
			&cli.BoolFlag{
				Name: "plain",
			},
		},
		Action: func(c *cli.Context) error {
			return BlockWorkspaceAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, workspace types.Workspace) error {
				client := api.Client{Config: cfg}
				showSensitive := c.IsSet("sensitive")
				outputs, err := client.WorkspaceOutputs().GetCurrent(stack.Id, workspace.Uid, showSensitive)
				if err != nil {
					return err
				}

				for key, output := range outputs {
					if output.Redacted {
						output.Value = "(hidden)"
						outputs[key] = output
					}
				}

				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if c.IsSet("plain") {
					stripped := map[string]any{}
					for key, output := range outputs {
						stripped[key] = output.Value
					}
					encoder.Encode(stripped)
				} else {
					encoder.Encode(outputs)
				}

				return nil
			})
		},
	}
}
