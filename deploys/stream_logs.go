package deploys

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"os"
	"time"
)

// StreamLogs streams the logs from the server over a websocket
// The logs are emitted to stdout
func StreamLogs(ctx context.Context, cfg api.Config, deploy *types.Deploy) error {
	// innerCtx will allow us to cancel when the deploy reaches a terminal status
	innerCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	fmt.Fprintln(os.Stdout, "Waiting for logs...")
	client := api.Client{Config: cfg}
	msgs, err := client.DeployLiveLogs().Watch(innerCtx, deploy.StackId, deploy.Id)
	if err != nil {
		return err
	}
	// NOTE: pollDeploy is needed to know when the live logs are complete
	// TODO: Replace pollDeploy with an EOF message received through the live logs
	deployCh := pollDeploy(innerCtx, cfg, deploy, time.Second)
	for {
		select {
		case msg := <-msgs:
			if msg.Source != "error" {
				fmt.Fprint(os.Stdout, msg.Content)
			}
		case dep := <-deployCh:
			switch dep.Status {
			case types.DeployStatusCancelled:
				return fmt.Errorf("Deploy was cancelled.")
			case types.DeployStatusCompleted:
				return nil
			case types.DeployStatusFailed:
				return fmt.Errorf("Deploy failed to complete: %s", dep.StatusMessage)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// pollDeploy returns a channel that repeatedly delivers details about the run separated by pollDelay
// If input ctx is cancelled, the returned channel is closed
func pollDeploy(ctx context.Context, cfg api.Config, deploy *types.Deploy, pollDelay time.Duration) <-chan types.Deploy {
	ch := make(chan types.Deploy)
	client := api.Client{Config: cfg}
	go func() {
		defer close(ch)
		for {
			run, _ := client.Deploys().Get(deploy.StackId, deploy.AppId, deploy.EnvId, deploy.Id)
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
