package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
)

var WaitForLaunchFlag = &cli.BoolFlag{
	Name: "wait-for-launch",
	Usage: `If a workspace is pending launch, this command will track the launch and execute after the launch completes successfully.
If the workspace launch fails, the command will fail before executing.
By default, this is enabled. Set wait-for-launch=false to disable.`,
	Value: true,
}

func WaitForLaunch(ctx context.Context, osWriters logging.OsWriters, cfg api.Config, appDetails app.Details, waitForLaunch bool) error {
	ws := *appDetails.Workspace
	if ws.Status == types.WorkspaceStatusProvisioned {
		return nil
	}
	if !waitForLaunch {
		return fmt.Errorf("app %q has not been provisioned in %q environment yet", appDetails.App.Name, appDetails.Env.Name)
	}

	// If we made it here, we want to wait for a launch and the workspace has not been provisioned yet
	// Let's look for all the runs on this workspace to see if there are any pending
	// If all runs we retrieve are terminal, we use the latest to determine whether we can proceed
	ntRuns, err := find.NonTerminalRuns(cfg, appDetails.App.StackId, ws.Uid)
	if err != nil {
		return err
	}
	if len(ntRuns) <= 0 {
		return fmt.Errorf("app %q has not been provisioned in %q environment yet", appDetails.App.Name, appDetails.Env.Name)
	}

	track := ntRuns[0]

	stderr := osWriters.Stderr()
	fmt.Fprintf(stderr, "Waiting for app %q to launch in %q environment...\n", appDetails.App.Name, appDetails.Env.Name)
	fmt.Fprintf(stderr, "Watching run for launch: %s\n", runs.GetBrowserUrl(cfg, ws, track))

	result, err := runs.WaitForTerminal(ctx, osWriters, cfg, ws, track)
	if err != nil {
		return err
	} else if result.Status == types.RunStatusCompleted {
		fmt.Fprintln(stderr, "App launched successfully.")
		fmt.Fprintln(stderr, "")
		return nil
	}
	fmt.Fprintf(stderr, "App failed to launch because run finished with %q status.\n", result.Status)
	return fmt.Errorf("Could not run command because app failed to launch")
}
