package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/go-api-client.v0/ws"
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
			updated, err := client.IntentWorkflows().Get(ctx, iw.StackId, iw.Id)
			if err != nil {
				return cur, fmt.Errorf("error waiting for deployment: %w", err)
			} else if updated != nil {
				cur = *updated
			}
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

// waitForIacSync watches an IaC sync intent workflow until it reaches a terminal state,
// logging top-level status transitions to w. The returned error encodes the exit code:
// completed -> nil (0), failed -> cli.Exit(.., 1), cancelled -> cli.Exit(.., 2),
// timeout -> cli.Exit(.., 3). A user interrupt surfaces as context.Canceled, which main.go
// maps to a clean exit 0.
func waitForIacSync(ctx context.Context, cfg api.Config, w io.Writer, iw types.IntentWorkflow, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := api.Client{Config: cfg}
	initial, ch, err := client.IntentWorkflows().WatchGet(ctx, iw.StackId, iw.Id, ws.RetryInfinite(time.Second))
	if err != nil {
		return iacSyncWatchErr(ctx, err, timeout)
	} else if initial == nil {
		return iacSyncWatchErr(ctx, nil, timeout)
	}

	cur := *initial
	last := types.IntentWorkflowStatus("")
	for {
		if cur.Status != last {
			if cur.StatusMessage != "" {
				fmt.Fprintf(w, "IaC sync %s: %s\n", cur.Status, cur.StatusMessage)
			} else {
				fmt.Fprintf(w, "IaC sync %s\n", cur.Status)
			}
			last = cur.Status
		}

		switch cur.Status {
		case types.IntentWorkflowStatusCompleted:
			fmt.Fprintln(w, "IaC sync completed successfully.")
			return nil
		case types.IntentWorkflowStatusFailed:
			return cli.Exit(fmt.Sprintf("IaC sync failed: %s", cur.StatusMessage), 1)
		case types.IntentWorkflowStatusCancelled:
			return cli.Exit("IaC sync was cancelled", 2)
		}

		select {
		case so, ok := <-ch:
			if !ok {
				return iacSyncWatchErr(ctx, nil, timeout)
			}
			if so.Err != nil {
				return iacSyncWatchErr(ctx, so.Err, timeout)
			}
			cur = so.Object.ApplyTo(cur)
		case <-ctx.Done():
			return iacSyncWatchErr(ctx, ctx.Err(), timeout)
		}
	}
}

// iacSyncWatchErr maps a watch failure to the right exit behavior: a deadline -> exit 3,
// a user interrupt -> context.Canceled (main.go exits 0), anything else -> exit 1.
func iacSyncWatchErr(ctx context.Context, cause error, timeout time.Duration) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return cli.Exit(fmt.Sprintf("Timed out after %s waiting for IaC sync to complete", timeout), 3)
	}
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(cause, context.Canceled) {
		return context.Canceled
	}
	if cause != nil {
		return cli.Exit(fmt.Sprintf("error watching IaC sync workflow: %s", cause), 1)
	}
	return cli.Exit("IaC sync watch ended unexpectedly", 1)
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
