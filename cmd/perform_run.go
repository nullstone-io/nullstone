package cmd

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	api_runs "gopkg.in/nullstone-io/go-api-client.v0/runs"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app_urls"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
	"os"
)

type PerformRunInput struct {
	Workspace  types.Workspace
	CommitSha  string
	IsApproved *bool
	IsDestroy  bool
	BlockType  types.BlockType
	StreamLogs bool
}

func PerformRun(ctx context.Context, cfg api.Config, input PerformRunInput) error {
	result, err := api_runs.Create(ctx, cfg, input.Workspace, input.CommitSha, input.IsApproved, input.IsDestroy, "")
	if err != nil {
		return fmt.Errorf("error creating run: %w", err)
	} else if result == nil {
		return fmt.Errorf("unable to create run")
	}

	var newRun types.Run
	if result.IntentWorkflow != nil {
		// When creating runs, we should have a primary workflow already
		pw := result.IntentWorkflow.PrimaryWorkflow
		if pw == nil {
			return fmt.Errorf("no primary workflow found")
		}
		fmt.Fprintf(os.Stdout, "created workflow run (id = %d)\n", pw.Id)
		fmt.Fprintln(os.Stdout, app_urls.GetWorkspaceWorkflow(cfg, *pw, input.BlockType == types.BlockTypeApplication))
		if newRun, err = waitForWorkspaceWorkflowRun(ctx, cfg, *pw); err != nil {
			return fmt.Errorf("error waiting for workflow run: %w", err)
		}
	} else if result.Run != nil {
		newRun = *result.Run
		fmt.Fprintf(os.Stdout, "created run %q\n", newRun.Uid)
		fmt.Fprintln(os.Stdout, app_urls.GetRun(cfg, input.Workspace, newRun))
	} else {
		return fmt.Errorf("run was not created")
	}

	if input.StreamLogs {
		err := runs.StreamLogs(ctx, cfg, input.Workspace, &newRun)
		if errors.Is(err, runs.ErrRunDisapproved) {
			// Disapproved plans are expected, return no error
			return nil
		}
		return err
	}
	return nil
}
