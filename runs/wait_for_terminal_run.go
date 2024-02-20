package runs

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/logging"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"sync"
	"time"
)

func WaitForTerminalRun(ctx context.Context, osWriters logging.OsWriters, cfg api.Config, ws types.Workspace, track types.Run,
	timeout time.Duration, approvalTimeout time.Duration) (types.Run, error) {
	stderr := osWriters.Stderr()
	// ctx already contains cancellation for Ctrl+C, innerCtx allows us to cancel pollRun when timeout occurs
	innerCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	// updatedRun is returned to allow the caller to decision off the run status upon completion
	var updatedRun types.Run

	// This timer provides a hard timeout for entire wait operation
	// If it hits first, we print an error message and cancel innerCtx to stop pollRun
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	// This timer provides a timeout when we reach "needs-approval"
	// This timer starts with an extremely large timeout value, so it doesn't trigger;
	// When we reach "needs-approval", the timer is reset to the user-specified "approvalTimeout"
	approvalTimer := time.NewTimer(7 * 24 * time.Hour)
	defer approvalTimer.Stop()
	var printApprovalMsg sync.Once
	onNeedsApproval := func() {
		printApprovalMsg.Do(func() {
			fmt.Fprintln(stderr, "Nullstone requires approval before applying infrastructure changes.")
			fmt.Fprintln(stderr, "Visit the infrastructure logs in a browser to approve/reject.")
			fmt.Fprintln(stderr, GetBrowserUrl(cfg, ws, updatedRun))
		})
		approvalTimer.Reset(approvalTimeout)
	}

	runCh := pollRun(innerCtx, cfg, ws.StackId, track.Uid, time.Second)
	for {
		select {
		case updatedRun = <-runCh:
			if types.IsTerminalRunStatus(updatedRun.Status) {
				return updatedRun, nil
			}
			if updatedRun.Status == types.RunStatusNeedsApproval {
				onNeedsApproval()
			}
		case <-timeoutTimer.C:
			fmt.Fprintln(stderr, "Timed out waiting for workspace to provision.")
			return updatedRun, fmt.Errorf("Operation cancelled waiting for provision")
		case <-approvalTimer.C:
			fmt.Fprintln(stderr, "Timed out waiting for workspace to be approved.")
			return updatedRun, fmt.Errorf("Operation cancelled waiting for approval")
		case <-ctx.Done():
			return updatedRun, fmt.Errorf("User cancelled operation")
		}
	}
}
