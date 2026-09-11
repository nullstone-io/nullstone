package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var WorkflowsCancel = &cli.Command{
	Name: "cancel",
	Description: `Cancels an in-progress workflow for a block in an environment.
Find the workflow ID with "nullstone workflows list --active".

The workflow moves to "cancelling" immediately. A running terraform/opentofu or docker process is interrupted
so it can stop cleanly; if it has not exited after the hard-stop timeout (5 minutes by default), it is killed.
The workflow reaches "cancelled" once every run, build, and deploy it owns has stopped.
A deploy whose rollout was already handed to the platform is not rolled back; Nullstone only stops watching it.
Cancelling a workflow that already finished is a no-op.`,
	Usage:     "Cancel an in-progress workflow",
	UsageText: "nullstone workflows cancel <workflow-id> [--stack=<stack-name>] --block=<block-name> --env=<env-name>",
	Flags: []cli.Flag{
		StackFlag,
		BlockFlag,
		EnvFlag,
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			if c.NArg() != 1 {
				return fmt.Errorf("usage: nullstone workflows cancel <workflow-id> --block=<block-name> --env=<env-name>")
			}
			workflowId, err := strconv.ParseInt(c.Args().First(), 10, 64)
			if err != nil {
				return fmt.Errorf("invalid workflow id %q: must be a number from `nullstone workflows list`", c.Args().First())
			}

			sbe, err := find.StackBlockEnvByName(ctx, cfg, c.String(StackFlag.Name), c.String(BlockFlag.Name), c.String(EnvFlag.Name))
			if err != nil {
				return err
			}

			client := api.Client{Config: cfg}
			workflow, err := client.WorkspaceWorkflows().Cancel(ctx, sbe.Stack.Id, sbe.Block.Id, sbe.Env.Id, workflowId)
			if err != nil {
				return fmt.Errorf("error cancelling workflow: %w", err)
			} else if workflow == nil {
				return fmt.Errorf("workflow %d not found", workflowId)
			}

			switch workflow.Status {
			case types.WorkspaceWorkflowStatusCancelling:
				fmt.Fprintf(c.App.Writer, "Workflow %d (%s) is cancelling. It will be cancelled once every run, build, and deploy has stopped.\n", workflow.Id, workflow.FriendlyAction)
			case types.WorkspaceWorkflowStatusCancelled:
				fmt.Fprintf(c.App.Writer, "Workflow %d (%s) is cancelled.\n", workflow.Id, workflow.FriendlyAction)
			default:
				fmt.Fprintf(c.App.Writer, "Workflow %d (%s) already finished with status %q; nothing to cancel.\n", workflow.Id, workflow.FriendlyAction, workflow.Status)
			}
			return nil
		})
	},
}
