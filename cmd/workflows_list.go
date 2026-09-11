package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var (
	WorkflowActiveFlag = &cli.BoolFlag{
		Name:  "active",
		Usage: "Only show workflows that are still in progress (queued, awaiting dependencies, needing approval, running, or cancelling).",
	}
	WorkflowStatusFlag = &cli.StringSliceFlag{
		Name: "status",
		Usage: `Only show workflows with this status. Can be specified multiple times.
       Statuses: queued, awaiting-dependencies, needs-approval, running, cancelling, completed, failed, cancelled`,
	}
	WorkflowLimitFlag = &cli.IntFlag{
		Name:  "limit",
		Value: 20,
		Usage: "Maximum number of workflows to show (newest first).",
	}
)

var WorkflowsList = &cli.Command{
	Name: "list",
	Description: `Shows the most recent workflows for a block in an environment, newest first.
Each workflow bundles the infrastructure run, build, and deploy performed together (e.g. an "update-deploy").
Use --active to find a workflow to cancel, or --status to filter by any set of statuses.`,
	Usage:     "List workflows for a block in an environment",
	UsageText: "nullstone workflows list [--stack=<stack-name>] --block=<block-name> --env=<env-name> [--active] [--status=<status>] [--limit=<n>]",
	Flags: []cli.Flag{
		StackFlag,
		BlockFlag,
		EnvFlag,
		WorkflowActiveFlag,
		WorkflowStatusFlag,
		WorkflowLimitFlag,
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			statuses, err := parseWorkflowStatusFilters(c)
			if err != nil {
				return err
			}

			sbe, err := find.StackBlockEnvByName(ctx, cfg, c.String(StackFlag.Name), c.String(BlockFlag.Name), c.String(EnvFlag.Name))
			if err != nil {
				return err
			}

			client := api.Client{Config: cfg}
			workflows, _, err := client.WorkspaceWorkflows().Query(ctx, sbe.Stack.Id, sbe.Block.Id, sbe.Env.Id, api.QueryWorkspaceWorkflowsInput{
				Page:     1,
				PerPage:  c.Int(WorkflowLimitFlag.Name),
				Statuses: statuses,
			})
			if err != nil {
				return fmt.Errorf("error listing workflows: %w", err)
			}

			buffer := TableBuffer{}
			buffer.AddFields("ID", "Action", "Status", "Started", "By", "Message")
			for _, workflow := range workflows {
				buffer.AddRow(map[string]interface{}{
					"ID":      workflow.Id,
					"Action":  workflow.FriendlyAction,
					"Status":  workflow.Status,
					"Started": workflow.CreatedAt.Local().Format("2006-01-02 15:04:05"),
					"By":      workflow.CreatedBy,
					"Message": workflow.StatusMessage,
				})
			}
			fmt.Println(buffer.String())
			return nil
		})
	},
}

// parseWorkflowStatusFilters turns --active/--status into the status filter sent to the API.
// --active expands to every non-terminal status; --status values are validated against the known statuses.
func parseWorkflowStatusFilters(c *cli.Context) ([]types.WorkspaceWorkflowStatus, error) {
	statuses := make([]types.WorkspaceWorkflowStatus, 0)
	if c.Bool(WorkflowActiveFlag.Name) {
		statuses = append(statuses, types.ActiveWorkspaceWorkflowStatuses...)
	}
	for _, raw := range c.StringSlice(WorkflowStatusFlag.Name) {
		status := types.WorkspaceWorkflowStatus(strings.ToLower(strings.TrimSpace(raw)))
		if !isKnownWorkflowStatus(status) {
			return nil, fmt.Errorf("invalid --status %q: must be one of %s", raw, strings.Join(knownWorkflowStatuses(), ", "))
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func knownWorkflowStatuses() []string {
	return []string{
		string(types.WorkspaceWorkflowStatusQueued),
		string(types.WorkspaceWorkflowStatusAwaiting),
		string(types.WorkspaceWorkflowStatusNeedsApproval),
		string(types.WorkspaceWorkflowStatusRunning),
		string(types.WorkspaceWorkflowStatusCancelling),
		string(types.WorkspaceWorkflowStatusCompleted),
		string(types.WorkspaceWorkflowStatusFailed),
		string(types.WorkspaceWorkflowStatusCancelled),
	}
}

func isKnownWorkflowStatus(status types.WorkspaceWorkflowStatus) bool {
	for _, known := range knownWorkflowStatuses() {
		if string(status) == known {
			return true
		}
	}
	return false
}
