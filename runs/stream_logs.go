package runs

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"os"
	"time"
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
	msgs, err := client.LiveLogs().Watch(innerCtx, workspace.StackId, newRun.Uid)
	if err != nil {
		return err
	}
	runCh := pollRun(innerCtx, cfg, workspace.StackId, newRun.Uid, time.Second)
	for {
		select {
		case msg := <-msgs:
			if msg.Source != "error" {
				fmt.Fprint(os.Stdout, msg.Content)
			}
		case run := <-runCh:
			if types.IsTerminalRunStatus(run.Status) {
				cancelFn()
				return nil
			}
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
