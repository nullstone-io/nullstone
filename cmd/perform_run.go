package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/colorstring"
	"gopkg.in/nullstone-io/go-api-client.v0"
	api_runs "gopkg.in/nullstone-io/go-api-client.v0/runs"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app_urls"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
)

type PerformRunInput struct {
	Workspace  types.Workspace
	CommitSha  string
	IsApproved *bool
	IsDestroy  bool
	BlockType  types.BlockType
	StreamLogs bool
}

func PerformRun(ctx context.Context, cfg api.Config, logger *log.Logger, input PerformRunInput) error {
	logger.Println("Performing run...")
	logger.SetPrefix("    ")
	defer logger.SetPrefix("")

	latestUpdateAt := time.Now().Add(time.Second)
	result, err := api_runs.Create(ctx, cfg, input.Workspace, input.CommitSha, input.IsApproved, latestUpdateAt, input.IsDestroy, "")
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
		logger.Println("Created workflow run")
		logger.Println(fmt.Sprintf("URL: %s", app_urls.GetWorkspaceWorkflow(cfg, *pw, input.BlockType == types.BlockTypeApplication)))
		if newRun, err = waitForWorkspaceWorkflowRun(ctx, cfg, *pw); err != nil {
			return fmt.Errorf("error waiting for workflow run: %w", err)
		}
	} else if result.Run != nil {
		newRun = *result.Run
		logger.Println("Created run")
		logger.Println(fmt.Sprintf("URL: %s", app_urls.GetRun(cfg, input.Workspace, newRun)))
	} else {
		return fmt.Errorf("run was not created")
	}

	if input.StreamLogs {
		err := runs.StreamLogs(ctx, cfg, logger, input.Workspace, &newRun)
		if errors.Is(err, runs.ErrRunDisapproved) {
			// Disapproved plans are expected, return no error
			return nil
		}
		return err
	}
	colorstring.Fprintln(logger.Writer(), "[green]Run started")
	return nil
}
