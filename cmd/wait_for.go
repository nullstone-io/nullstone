package cmd

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/go-api-client.v0/ws"
	"time"
)

func waitForRunningIntentWorkflow(ctx context.Context, cfg api.Config, iw types.IntentWorkflow) (types.IntentWorkflow, error) {
	client := api.Client{Config: cfg}
	intentWorkflow, ch, err := client.IntentWorkflows().WatchGet(ctx, iw.StackId, iw.Id, ws.RetryInfinite(time.Second))
	if err != nil {
		return iw, fmt.Errorf("error waiting for deployment: %w", err)
	} else if intentWorkflow == nil {
		return iw, context.Canceled
	}

	cur := *intentWorkflow
	for {
		switch cur.Status {
		case types.IntentWorkflowStatusRunning:
			return cur, nil
		case types.IntentWorkflowStatusCompleted:
			return cur, nil
		case types.IntentWorkflowStatusFailed:
			return cur, fmt.Errorf("Deployment failed: %s", cur.StatusMessage)
		case types.IntentWorkflowStatusCancelled:
			return cur, fmt.Errorf("Deployment was cancelled.")
		}
		so := <-ch
		if so.Err != nil {
			return cur, fmt.Errorf("error waiting for deployment: %w", so.Err)
		}
		cur = so.Object.ApplyTo(cur)
	}
}

func waitForCompletedIntentWorkflow(ctx context.Context, cfg api.Config, iw types.IntentWorkflow) (types.IntentWorkflow, error) {
	client := api.Client{Config: cfg}
	intentWorkflow, ch, err := client.IntentWorkflows().WatchGet(ctx, iw.StackId, iw.Id, ws.RetryInfinite(time.Second))
	if err != nil {
		return iw, fmt.Errorf("error waiting for deployment: %w", err)
	} else if intentWorkflow == nil {
		return iw, context.Canceled
	}

	cur := *intentWorkflow
	for {
		switch cur.Status {
		case types.IntentWorkflowStatusCompleted:
			return cur, nil
		case types.IntentWorkflowStatusFailed:
			return cur, fmt.Errorf("Deployment failed: %s", cur.StatusMessage)
		case types.IntentWorkflowStatusCancelled:
			return cur, fmt.Errorf("Deployment was cancelled.")
		}
		so := <-ch
		if so.Err != nil {
			return cur, fmt.Errorf("error waiting for deployment: %w", so.Err)
		}
		cur = so.Object.ApplyTo(cur)
	}
}

func waitForWorkspaceWorkflowRun(ctx context.Context, cfg api.Config, ww types.WorkspaceWorkflow) (types.Run, error) {
	client := api.Client{Config: cfg}
	workspaceWorkflow, ch, err := client.WorkspaceWorkflows().WatchGet(ctx, ww.StackId, ww.BlockId, ww.EnvId, ww.Id, ws.RetryInfinite(time.Second))
	if err != nil {
		return types.Run{}, fmt.Errorf("error waiting for run: %w", err)
	} else if workspaceWorkflow == nil {
		return types.Run{}, context.Canceled
	}

	cur := *workspaceWorkflow
	for {
		if cur.Run != nil {
			return *cur.Run, nil
		}
		if types.IsTerminalWorkspaceWorkflow(cur.Status) {
			return types.Run{}, fmt.Errorf("workflow reached %s status before a run could be found", cur.Status)
		}
		so := <-ch
		if so.Err != nil {
			return types.Run{}, fmt.Errorf("error waiting for run: %w", so.Err)
		}
		cur = so.Object.ApplyTo(cur)
	}
}
