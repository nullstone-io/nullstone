package runs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/colorstring"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/go-api-client.v0/ws"
	"gopkg.in/nullstone-io/nullstone.v0/app_urls"
)

var (
	ErrRunDisapproved = errors.New("run was disapproved")
)

// RunFailedError is returned when a run reaches a terminal failure status.
// It includes the run's phase and status for callers that need to distinguish
// between plan failures and apply failures.
type RunFailedError struct {
	Phase         string
	Status        string
	StatusMessage string
}

func (e *RunFailedError) Error() string {
	return fmt.Sprintf("run failed to complete (%s): %s", e.Status, e.StatusMessage)
}

// StreamLogs streams the logs from the server over a websocket
// The logs are emitted to stdout
func StreamLogs(ctx context.Context, cfg api.Config, logger *log.Logger, workspace types.Workspace, newRun *types.Run) error {
	// ctx already contains cancellation for Ctrl+C
	// innerCtx will allow us to cancel when the run reaches a terminal status
	innerCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	logger.Println("Streaming run logs...")
	logger.Println()
	logger.Println()
	client := api.Client{Config: cfg}
	msgs, err := client.RunLogs().Watch(innerCtx, workspace.StackId, newRun.Uid, ws.RetryInfinite(1*time.Second))
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
				return drainAndFinalize(ctx, logger, msgs, run)
			}
			if run.Status == types.RunStatusNeedsApproval {
				printApprovalMsg.Do(func() {
					logger.Println("Nullstone requires approval before applying infrastructure changes.")
					logger.Println("Visit the infrastructure logs in a browser to approve/reject.")
					logger.Println(fmt.Sprintf("URL: %s", app_urls.GetRun(cfg, workspace, run)))
				})
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// drainAndFinalize drains remaining log messages after a run reaches a terminal status.
// It waits until 1.1s passes with no new log content before returning the final result.
func drainAndFinalize(ctx context.Context, logger *log.Logger, msgs <-chan types.Message, run types.Run) error {
	flushTimeout := 1100 * time.Millisecond
	timer := time.NewTimer(flushTimeout)
	defer timer.Stop()
	for {
		select {
		case msg := <-msgs:
			if msg.Type != "error" {
				fmt.Fprint(os.Stdout, msg.Content)
			}
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(flushTimeout)
		case <-timer.C:
			logger.Println()
			if run.Status == types.RunStatusDisapproved {
				colorstring.Fprintln(logger.Writer(), "[yellow]Run disapproved")
				return ErrRunDisapproved
			}
			if run.Status != types.RunStatusCompleted {
				colorstring.Fprintln(logger.Writer(), "[red]Run failed")
				return &RunFailedError{
					Phase:         run.Phase,
					Status:        run.Status,
					StatusMessage: run.StatusMessage,
				}
			}
			colorstring.Fprintln(logger.Writer(), "[green]Run completed")
			return nil
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
			run, _ := client.Runs().Get(ctx, stackId, runUid)
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
