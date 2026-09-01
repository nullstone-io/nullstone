package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/workspace"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app_urls"
	"time"
)

var waitForDeployedPollDelay = 2 * time.Second

// WaitForDeployed waits for an app to reach a successful deployment in an environment.
// Launches and deploys can start out-of-band (CI, the UI, another terminal), so this must tolerate
// any starting condition: not provisioned yet, launch in progress, deploy triggered but the deploy
// record not created yet, deploy in flight, or already deployed.
//
// Each poll inspects the latest deploy and the newest workspace workflow:
//   - the latest deploy completed and no workflow started after it -> success
//   - anything else (no deploy yet, deploy in flight, failed/cancelled deploy that may be retried,
//     workflow running) -> keep waiting until a deploy succeeds or the timeout expires
func WaitForDeployed(ctx context.Context, osWriters logging.OsWriters, cfg api.Config, details workspace.Details, timeout time.Duration) error {
	stderr := osWriters.Stderr()
	if details.Block.Type != string(types.BlockTypeApplication) {
		return fmt.Errorf("The wait command with --for=deployed only supports apps; %q is a %s", details.Block.Name, details.Block.Type)
	}

	fmt.Fprintf(stderr, "Waiting for %q to deploy in %q environment...\n", details.Block.Name, details.Env.Name)
	fmt.Fprintf(stderr, "Timeout = %s\n", timeout)

	client := api.Client{Config: cfg}
	stackId, appId, envId := details.Block.StackId, details.Block.Id, details.Env.Id

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	var lastDeploy *types.Deploy
	firstCheck := true
	reportedQuiet := false
	reportedWorkflowId := int64(0)
	lastReported := ""
	for {
		deploy, err := client.Deploys().GetLatest(ctx, stackId, appId, envId, nil)
		if err != nil {
			return fmt.Errorf("error retrieving latest deploy: %w", err)
		}
		if deploy != nil {
			lastDeploy = deploy
		}

		wflows, _, err := client.WorkspaceWorkflows().List(ctx, stackId, appId, envId, 1, 1)
		if err != nil {
			return fmt.Errorf("error retrieving workspace workflows: %w", err)
		}
		var newestWorkflow *types.WorkspaceWorkflow
		if len(wflows) > 0 {
			newestWorkflow = &wflows[0]
		}
		hasActivity := newestWorkflow != nil && !types.IsTerminalWorkspaceWorkflow(newestWorkflow.Status)
		if hasActivity && newestWorkflow.Id != reportedWorkflowId {
			reportedWorkflowId = newestWorkflow.Id
			fmt.Fprintf(stderr, "Workflow %q is in progress: %s\n", newestWorkflow.FriendlyAction, app_urls.GetWorkspaceWorkflow(cfg, *newestWorkflow, true))
		}

		if deploy != nil {
			reported := fmt.Sprintf("%d:%s", deploy.Id, deploy.Status)
			switch deploy.Status {
			case types.DeployStatusCompleted:
				// A workflow created after this deploy is producing a newer deploy; this one is stale
				if !hasActivity || !newestWorkflow.CreatedAt.After(deploy.CreatedAt) {
					if firstCheck {
						fmt.Fprintln(stderr, "App has deployed already.")
					} else {
						fmt.Fprintln(stderr, "App deployed successfully.")
					}
					return nil
				}
			case types.DeployStatusFailed, types.DeployStatusCancelled:
				// A failed/cancelled deploy may be retried out-of-band; wait for new activity until timeout
				if reported != lastReported {
					fmt.Fprintf(stderr, "Deploy %s %s; waiting for a new deploy to start...\n", deployLabel(*deploy), deploy.Status)
				}
			default: // queued, initializing, running
				if reported != lastReported {
					fmt.Fprintf(stderr, "Deploy %s is %s...\n", deployLabel(*deploy), deploy.Status)
				}
			}
			lastReported = reported
		} else if !hasActivity && !reportedQuiet {
			reportedQuiet = true
			fmt.Fprintln(stderr, "No deploy has started yet. Waiting for a deploy to start...")
		}
		firstCheck = false

		select {
		case <-time.After(waitForDeployedPollDelay):
		case <-timeoutTimer.C:
			fmt.Fprintln(stderr, "Timed out waiting for app to deploy.")
			if lastDeploy != nil {
				switch lastDeploy.Status {
				case types.DeployStatusFailed, types.DeployStatusCancelled:
					return fmt.Errorf("Operation cancelled waiting for deploy; deploy %s finished with %q status: %s", deployLabel(*lastDeploy), lastDeploy.Status, lastDeploy.StatusMessage)
				}
			}
			return fmt.Errorf("Operation cancelled waiting for deploy")
		case <-ctx.Done():
			return fmt.Errorf("User cancelled operation")
		}
	}
}

func deployLabel(deploy types.Deploy) string {
	if deploy.Reference != "" {
		return deploy.Reference
	}
	return fmt.Sprintf("#%d", deploy.Id)
}
