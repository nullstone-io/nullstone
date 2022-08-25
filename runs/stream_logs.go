package runs

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/go-api-client.v0/ws"
	"os"
	"sync"
	"time"
)

var (
	ErrRunDisapproved = errors.New("run was disapproved")
)

// StreamLogs streams the logs from the server over a websocket
// The logs are emitted to stdout
func StreamLogs(ctx context.Context, cfg api.Config, workspace types.Workspace, newRun *types.Run) error {
	// ctx already contains cancellation for Ctrl+C
	// innerCtx will allow us to cancel when the run reaches a terminal status
	innerCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	fmt.Fprintln(os.Stdout, "Waiting for logs...")
	client := api.Client{Config: cfg}
	msgs, err := client.RunLogs().Watch(innerCtx, workspace.StackId, newRun.Uid, ws.RetryInfinite(time.Second))
	if err != nil {
		return err
	}
	// NOTE: pollRun is needed to know when the live logs are complete
	// TODO: Replace pollRun with an EOF message received through the live logs
	runCh := pollRun(innerCtx, cfg, workspace.StackId, newRun.Uid, time.Second)
	var printApprovalMsg sync.Once
	for {
		select {
		case msg := <-msgs:
			if msg.Type != "error" {
				fmt.Fprint(os.Stdout, msg.Content)
			}
		case run := <-runCh:
			if types.IsTerminalRunStatus(run.Status) {
				// A completed run finishes successfully
				// Any other terminal status returns an error (causing a non-zero exit code for failed runs)
				if run.Status == types.RunStatusDisapproved {
					return ErrRunDisapproved
				}
				if run.Status != types.RunStatusCompleted {
					return fmt.Errorf("Run failed to complete (%s): %s", run.Status, run.StatusMessage)
				}
				return nil
			}
			if run.Status == types.RunStatusNeedsApproval {
				printApprovalMsg.Do(func() {
					fmt.Fprintln(os.Stdout, "Nullstone requires approval before applying infrastructure changes.")
					fmt.Fprintln(os.Stdout, "Visit the infrastructure logs in a browser to approve/reject.")
					fmt.Fprintln(os.Stdout, GetBrowserUrl(cfg, workspace, run))
				})
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// pollRun returns a channel that repeatedly delivers details about the run separated by pollDelay
// If input ctx is cancelled, the returned channel is closed
func pollRun(ctx context.Context, cfg api.Config, stackId int64, runUid uuid.UUID, pollDelay time.Duration) <-chan types.Run {
	ch := make(chan types.Run)
	client := api.Client{Config: cfg}
	go func() {
		defer close(ch)
		for {
			run, _ := client.Runs().Get(stackId, runUid)
			if run != nil {
				ch <- *run
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(pollDelay):
			}
		}
	}()
	return ch
}
