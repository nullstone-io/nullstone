package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/workspace"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
	"strings"
	"time"
)

var (
	WaitForFlag = &cli.StringFlag{
		Name: "for",
		Usage: `Configure the wait command to reach a specific status. 
       Currently this supports --for=launched.
       In the future, we will support --for=destroyed and --for=deployed`,
	}
	WaitTimeoutFlag = &cli.DurationFlag{
		Name:        "timeout",
		DefaultText: "1s",
		Usage: `Set --timeout to a golang duration to control how long to wait for a status before cancelling.
       The default is '1h' (1 hour).
      `,
	}
	WaitApprovalTimeout = &cli.DurationFlag{
		Name:        "approval-timeout",
		DefaultText: "1s",
		Usage: `Set --approval-timeout to a golang duration to control how long to wait for approval before cancelling.
       If the workspace run never reaches "needs-approval", this has no effect.
       The default is '15m' (15 minutes).
      `,
	}
)

var Wait = func() *cli.Command {
	return &cli.Command{
		Name: "wait",
		Description: `Waits for a workspace to reach a specific status.
This is helpful to wait for infrastructure to provision or an app to deploy.
Currently, this supports --for=launched to wait for a workspace to provision.
In the future, we will add --for=destroyed and --for=deployed.`,
		Usage:     "Wait for a block to launch, destroy, or deploy in an environment.",
		UsageText: "nullstone wait [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			BlockFlag,
			EnvFlag,
			WaitForFlag,
		},
		Action: func(c *cli.Context) error {
			waitFor := c.String(WaitForFlag.Name)
			timeout := c.Duration(WaitTimeoutFlag.Name)
			approvalTimeout := c.Duration(WaitApprovalTimeout.Name)

			return BlockWorkspaceAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, ws types.Workspace) error {
				details := workspace.Details{
					Block:     &block,
					Env:       &env,
					Workspace: &ws,
					Module:    nil,
				}
				osWriters := CliOsWriters{Context: c}
				switch strings.ToLower(waitFor) {
				case "launched":
					return WaitForLaunch(ctx, osWriters, cfg, details, timeout, approvalTimeout)
				default:
					return fmt.Errorf("The wait command does not support --for=%s", waitFor)
				}
			})
		},
	}
}

func WaitForLaunch(ctx context.Context, osWriters logging.OsWriters, cfg api.Config, details workspace.Details,
	timeout time.Duration, approvalTimeout time.Duration) error {
	if details.Workspace.Status == types.WorkspaceStatusProvisioned {
		return nil
	}

	// If we made it here, we want to wait for a launch and the workspace has not been provisioned yet
	// Let's look for all the runs on this workspace to see if there are any pending
	// If we retrieve non-terminal runs, we use the oldest (next in line) to determine whether we can proceed
	launchRun, err := findLaunchRun(cfg, details)
	if err != nil {
		return err
	} else if launchRun == nil {
		return fmt.Errorf("app %q has not been provisioned in %q environment yet", details.Block.Name, details.Env.Name)
	}

	stderr := osWriters.Stderr()
	fmt.Fprintf(stderr, "Waiting for app %q to launch in %q environment...\n", details.Block.Name, details.Env.Name)
	fmt.Fprintf(stderr, "Watching run for launch: %s\n", runs.GetBrowserUrl(cfg, *details.Workspace, *launchRun))

	result, err := runs.WaitForTerminalRun(ctx, osWriters, cfg, *details.Workspace, *launchRun, timeout, approvalTimeout)
	if err != nil {
		return err
	} else if result.Status == types.RunStatusCompleted {
		fmt.Fprintln(stderr, "App launched successfully.")
		fmt.Fprintln(stderr, "")
	}
	fmt.Fprintf(stderr, "App failed to launch because run finished with %q status.\n", result.Status)
	return fmt.Errorf("Could not run command because app failed to launch")
}

func findLaunchRun(cfg api.Config, details workspace.Details) (*types.Run, error) {
	ntRuns, err := find.NonTerminalRuns(cfg, details.Block.StackId, details.Workspace.Uid)
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
