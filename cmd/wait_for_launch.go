package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
)

func WaitForLaunch(ctx context.Context, osWriters logging.OsWriters, cfg api.Config, appDetails *app.Details, waitForLaunch bool) error {
	ws := *appDetails.Workspace
	if ws.Status == types.WorkspaceStatusProvisioned {
		return nil
	}
	if !waitForLaunch {
		return fmt.Errorf("app %q has not been provisioned in %q environment yet", appDetails.App.Name, appDetails.Env.Name)
	}

	// If we made it here, we want to wait for a launch and the workspace has not been provisioned yet
	// Let's look for all the runs on this workspace to see if there are any pending
	// If we retrieve non-terminal runs, we use the oldest (next in line) to determine whether we can proceed
	launchRun, err := findLaunchRun(cfg, *appDetails)
	if err != nil {
		return err
	} else if launchRun == nil {
		return fmt.Errorf("app %q has not been provisioned in %q environment yet", appDetails.App.Name, appDetails.Env.Name)
	}

	stderr := osWriters.Stderr()
	fmt.Fprintf(stderr, "Waiting for app %q to launch in %q environment...\n", appDetails.App.Name, appDetails.Env.Name)
	fmt.Fprintf(stderr, "Watching run for launch: %s\n", runs.GetBrowserUrl(cfg, ws, *launchRun))

	result, err := runs.WaitForTerminal(ctx, osWriters, cfg, ws, *launchRun)
	if err != nil {
		return err
	} else if result.Status == types.RunStatusCompleted {
		fmt.Fprintln(stderr, "App launched successfully.")
		fmt.Fprintln(stderr, "")

		// We need to reload the workspace to hydrate with the LastFinishedRun
		client := api.Client{Config: cfg}
		appDetails.Workspace, err = client.Workspaces().Get(appDetails.App.StackId, appDetails.App.Id, appDetails.Env.Id)
		if err != nil {
			return fmt.Errorf("The app launched, but there was an error retrieving the workspace: %w", err)
		}
		return nil
	}
	fmt.Fprintf(stderr, "App failed to launch because run finished with %q status.\n", result.Status)
	return fmt.Errorf("Could not run command because app failed to launch")
}

func findLaunchRun(cfg api.Config, appDetails app.Details) (*types.Run, error) {
	ntRuns, err := find.NonTerminalRuns(cfg, appDetails.App.StackId, appDetails.Workspace.Uid)
	if err != nil {
		return nil, err
	}

	// It's possible that the queue of runs contains a "destroy" run
	// Let's find the first that is a launch/apply
	for _, run := range ntRuns {
		if run.IsDestroy {
			continue
		}
		return &run, nil
	}
	return nil, nil
}
