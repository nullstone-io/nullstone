package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/go-api-client.v0/ws"
	"gopkg.in/nullstone-io/nullstone.v0/artifacts"
)

var Deploy = func(providers app.Providers) *cli.Command {
	waitFlag := &cli.BoolFlag{
		Name:    "wait",
		Aliases: []string{"w"},
		Usage:   "Wait for the deploy to complete and stream the logs to the console.",
	}

	return &cli.Command{
		Name:        "deploy",
		Description: "Deploy a new version of your code for this application. This command works in tandem with the `nullstone push` command. This command deploys the artifacts that were uploaded during the `push` command.",
		Usage:       "Deploy application",
		UsageText:   "nullstone deploy [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppVersionFlag,
			waitFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				osWriters := CliOsWriters{Context: c}
				wait := c.IsSet(waitFlag.Name)

				info, err := calcDeployInfo(ctx, c)
				if err != nil {
					return err
				}

				fmt.Fprintln(osWriters.Stderr(), "Creating deploy...")
				result, err := CreateDeploy(cfg, appDetails, info)
				if err != nil {
					return err
				}
				fmt.Fprintln(osWriters.Stderr())

				if result.Deploy != nil {
					return streamDeployLogs(ctx, osWriters, cfg, *result.Deploy, wait)
				} else if result.IntentWorkflow != nil {
					return streamDeployIntentLogs(ctx, osWriters, cfg, appDetails, *result.IntentWorkflow, wait)
				}
				return nil
			})
		},
	}
}

func calcDeployInfo(ctx context.Context, c *cli.Context) (artifacts.VersionInfo, error) {
	osWriters := CliOsWriters{Context: c}
	stderr := osWriters.Stderr()
	version := c.String(AppVersionFlag.Name)

	if version == "" {
		fmt.Fprintf(stderr, "No version specified. Defaulting version based on current git commit sha...\n")
	}
	info, err := artifacts.GetVersionInfoFromWorkingDir(version)
	if err != nil {
		return info, err
	}
	if version == "" {
		fmt.Fprintf(stderr, "Version defaulted to %q\n", info.DesiredVersion)
	}

	return info, nil
}

func streamDeployLogs(ctx context.Context, osWriters logging.OsWriters, cfg api.Config, deploy types.Deploy, wait bool) error {
	stdout, stderr := osWriters.Stdout(), osWriters.Stderr()
	client := api.Client{Config: cfg}

	fmt.Fprintln(stderr, "Waiting for deploy logs...")
	msgs, err := client.DeployLogs().Watch(ctx, deploy.StackId, deploy.Id, ws.RetryInfinite(2*time.Second))
	if err != nil {
		return fmt.Errorf("error connecting to deploy logs: %w", err)
	}
	for msg := range msgs {
		if msg.Type == "error" {
			return fmt.Errorf("%s", msg.Content)
		}
		if !wait && msg.Context == types.DeployPhaseWaitHealthy {
			// Stop streaming logs if we receive a log message from wait-healthy and no --wait
			break
		}
		fmt.Fprint(stderr, msg.Content)
	}

	updated, err := client.Deploys().Get(ctx, deploy.StackId, deploy.AppId, deploy.EnvId, deploy.Id)
	if err != nil {
		return fmt.Errorf("error retrieving deploy status: %w", err)
	}
	fmt.Fprintln(stdout, updated.Reference)
	switch updated.Status {
	case types.DeployStatusCompleted:
		colorstring.Fprintln(stderr, "[green]Deployment completed.[reset]")
	case types.DeployStatusCancelled:
		return fmt.Errorf("Deployment was cancelled.")
	case types.DeployStatusFailed:
		return fmt.Errorf("Deployment failed to complete: %s", updated.StatusMessage)
	}
	return nil
}

func streamDeployIntentLogs(ctx context.Context, osWriters logging.OsWriters, cfg api.Config, appDetails app.Details, iw types.IntentWorkflow, wait bool) error {
	_, stderr := osWriters.Stdout(), osWriters.Stderr()
	client := api.Client{Config: cfg}

	fmt.Fprintln(stderr, "Starting deployment...")
	var err error
	if iw, err = waitForRunningIntentWorkflow(ctx, cfg, iw); err != nil {
		return err
	} else if iw.Status == types.IntentWorkflowStatusCompleted {
		colorstring.Fprintln(stderr, "[green]Deployment completed.[reset]")
		return nil
	}

	colorstring.Fprintln(stderr, "[green]Deployment started.[reset]")

	var wflow types.WorkspaceWorkflow
	for _, ww := range iw.WorkspaceWorkflows {
		if ww.BlockId == appDetails.App.Id && ww.EnvId == appDetails.Env.Id && ww.StackId == appDetails.App.StackId {
			wflow = ww
			break
		}
	}
	if wflow.Id == 0 {
		colorstring.Fprintln(stderr, "[yellow]We cannot find the deployment workflow to report the deployment logs.[reset]")
		colorstring.Fprintln(stderr, "[bold]Waiting for successful deployment without reporting deployment logs...[reset]")
		iw, err = waitForCompletedIntentWorkflow(ctx, cfg, iw)
		if err == nil {
			colorstring.Fprintln(stderr, "[green]Deployment completed.[reset]")
		}
		return err
	}

	activities, err := client.WorkspaceWorkflows().GetActivities(ctx, wflow.StackId, wflow.BlockId, wflow.EnvId, wflow.Id)
	if err != nil {
		return fmt.Errorf("unable to find deployment: %w", err)
	} else if activities == nil || activities.Deploy == nil {
		return fmt.Errorf("deployment is missing")
	}

	return streamDeployLogs(ctx, osWriters, cfg, *activities.Deploy, wait)
}
