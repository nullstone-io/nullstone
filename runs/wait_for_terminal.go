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

var (
	WaitForTerminalTimeout      = time.Hour
	WaitForNeedsApprovalTimeout = 15 * time.Minute
)

func WaitForTerminal(ctx context.Context, osWriters logging.OsWriters, cfg api.Config, ws types.Workspace, track types.Run) types.Run {
	stderr := osWriters.Stderr()
	// ctx already contains cancellation for Ctrl+C
	// innerCtx will allow us to cancel when the run reaches a terminal status
	innerCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	var updatedRun types.Run
	var printApprovalMsg sync.Once

	// wait timeout starts off with a hard timeout for the entire operation
	// After reaching "needs-approval", this timer is reset to a soft timeout waiting for needs-approval
	waitTimeout := time.NewTimer(WaitForTerminalTimeout)
	defer waitTimeout.Stop()

	runCh := pollRun(innerCtx, cfg, ws.StackId, track.Uid, time.Second)
	for {
		select {
		case updatedRun = <-runCh:
			if types.IsTerminalRunStatus(updatedRun.Status) {
				return updatedRun
			}
			if updatedRun.Status == types.RunStatusNeedsApproval {
				printApprovalMsg.Do(func() {
					fmt.Fprintln(stderr, "Nullstone requires approval before applying infrastructure changes.")
					fmt.Fprintln(stderr, "Visit the infrastructure logs in a browser to approve/reject.")
					fmt.Fprintln(stderr, GetBrowserUrl(cfg, ws, updatedRun))
				})
				waitTimeout.Reset(WaitForNeedsApprovalTimeout)
			}
		case <-waitTimeout.C:
			fmt.Fprintln(stderr, "Timed out waiting for app to provision.")
			return updatedRun
		case <-ctx.Done():
			return updatedRun
		}
	}
}
