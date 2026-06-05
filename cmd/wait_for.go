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
		case types.IntentWorkflowStatusNoOp:
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

// isTerminalIntentWorkflow reports whether an intent workflow has reached a state where no further
// work (and therefore no deploy) will be produced.
func isTerminalIntentWorkflow(status types.IntentWorkflowStatus) bool {
	switch status {
	case types.IntentWorkflowStatusCompleted,
		types.IntentWorkflowStatusFailed,
		types.IntentWorkflowStatusCancelled,
		types.IntentWorkflowStatusNoOp:
		return true
	}
	return false
}

// waitForDeployActivity waits for the deploy belonging to wflow to be created and returns it.
//
// A release flips its intent workflow to "running" before the deploy row exists (the deploy is
// created later, inside the workspace workflow execution), so reading activities the instant we see
// "running" races the deploy's creation. This resolves when one of three things happens:
//  1. the deploy activity appears -> returned
//  2. the user cancels (ctx) -> context.Canceled (main.go maps this to a clean exit)
//  3. the intent workflow reaches a terminal status without ever producing a deploy -> error
//
// The deploy arrives on the workspace workflow stream while the terminal signal arrives on the intent
// workflow stream; because those are independent connections, we re-check activities before declaring
// the deploy missing so a deploy created in the same flush as the terminal status isn't lost.
func waitForDeployActivity(ctx context.Context, cfg api.Config, iw types.IntentWorkflow, wflow types.WorkspaceWorkflow) (*types.Deploy, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	client := api.Client{Config: cfg}

	// Fast path: the deploy may already exist (created before, or just as, the intent went running).
	if deploy, err := getDeployActivity(ctx, client, wflow); err != nil {
		return nil, err
	} else if deploy != nil {
		return deploy, nil
	}

	// Watch the workspace workflow for the deploy and the intent workflow for a terminal status.
	_, wwCh, err := client.WorkspaceWorkflows().WatchGet(ctx, wflow.StackId, wflow.BlockId, wflow.EnvId, wflow.Id, ws.RetryInfinite(time.Second))
	if err != nil {
		return nil, fmt.Errorf("error waiting for deployment: %w", err)
	}
	initialIw, iwCh, err := client.IntentWorkflows().WatchGet(ctx, iw.StackId, iw.Id, ws.RetryInfinite(time.Second))
	if err != nil {
		return nil, fmt.Errorf("error waiting for deployment: %w", err)
	}

	// Re-check now that we're subscribed: a deploy created between the fast-path check and our
	// subscription wouldn't appear on the stream (the hydrate snapshot carries no deploy).
	if deploy, err := getDeployActivity(ctx, client, wflow); err != nil {
		return nil, err
	} else if deploy != nil {
		return deploy, nil
	}

	// The intent workflow may already be terminal (e.g. a no-op release) by the time we connect.
	if initialIw != nil && isTerminalIntentWorkflow(initialIw.Status) {
		return deployOrMissing(ctx, client, wflow)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		case so, ok := <-wwCh:
			if !ok {
				return nil, context.Canceled
			}
			if so.Err != nil {
				return nil, fmt.Errorf("error waiting for deployment: %w", so.Err)
			}
			if so.Object.Deploy != nil {
				return so.Object.Deploy, nil
			}
		case so, ok := <-iwCh:
			if !ok {
				return nil, context.Canceled
			}
			if so.Err != nil {
				return nil, fmt.Errorf("error waiting for deployment: %w", so.Err)
			}
			if so.Object.Status != nil && isTerminalIntentWorkflow(*so.Object.Status) {
				return deployOrMissing(ctx, client, wflow)
			}
		}
	}
}

// getDeployActivity fetches the deploy attached to wflow, returning nil if one hasn't been created yet.
func getDeployActivity(ctx context.Context, client api.Client, wflow types.WorkspaceWorkflow) (*types.Deploy, error) {
	activities, err := client.WorkspaceWorkflows().GetActivities(ctx, wflow.StackId, wflow.BlockId, wflow.EnvId, wflow.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to find deployment: %w", err)
	}
	if activities == nil {
		return nil, nil
	}
	return activities.Deploy, nil
}

// deployOrMissing re-checks activities one last time (closing the cross-stream race) and reports the
// deploy if it landed, otherwise the canonical "deployment is missing" error.
func deployOrMissing(ctx context.Context, client api.Client, wflow types.WorkspaceWorkflow) (*types.Deploy, error) {
	if deploy, err := getDeployActivity(ctx, client, wflow); err != nil {
		return nil, err
	} else if deploy != nil {
		return deploy, nil
	}
	return nil, fmt.Errorf("deployment is missing")
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
		case types.IntentWorkflowStatusNoOp:
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
