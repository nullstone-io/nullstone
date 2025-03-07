package cmd

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app_urls"
	"os"
)

type PerformEnvRunInput struct {
	CommitSha string
	Stack     types.Stack
	Env       types.Environment
	IsDestroy bool
}

func PerformEnvRun(ctx context.Context, cfg api.Config, input PerformEnvRunInput) error {
	stdout := os.Stdout
	action := "launch"
	if input.IsDestroy {
		action = "destroy"
	}

	client := api.Client{Config: cfg}
	body := types.CreateEnvRunInput{IsDestroy: input.IsDestroy}
	result, err := client.EnvRuns().Create(ctx, input.Stack.Id, input.Env.Id, body)
	if err != nil {
		return fmt.Errorf("error creating run: %w", err)
	} else if result == nil {
		fmt.Fprintf(stdout, "no runs created to %s the %q environment\n", action, input.Env.Name)
		return nil
	}

	if result.IntentWorkflow.Intent != "" {
		fmt.Fprintf(stdout, "created workflow to %s %q environment.\n", action, input.Env.Name)
		fmt.Fprintln(stdout, app_urls.GetIntentWorkflow(cfg, result.IntentWorkflow))
		return nil
	} else if result.Runs == nil {
		return fmt.Errorf("workflow to %q environment was not created", action)
	}

	if len(result.Runs) < 1 {
		fmt.Fprintf(stdout, "no runs created to %s the %q environment\n", action, input.Env.Name)
		return nil
	}

	workspaces, err := client.Workspaces().List(ctx, input.Env.Id)
	if err != nil {
		return fmt.Errorf("error retrieving list of workspaces: %w", err)
	}
	blocks, err := client.Blocks().List(ctx, input.Stack.Id, false)
	if err != nil {
		return fmt.Errorf("error retrieving list of blocks: %w", err)
	}

	findWorkspace := func(run types.Run) *types.Workspace {
		for _, workspace := range workspaces {
			if workspace.Uid == run.WorkspaceUid {
				return &workspace
			}
		}
		return nil
	}
	findBlock := func(workspace *types.Workspace) *types.Block {
		if workspace == nil {
			return nil
		}
		for _, block := range blocks {
			if workspace.BlockId == block.Id {
				return &block
			}
		}
		return nil
	}
	for _, run := range result.Runs {
		blockName := "(unknown)"
		workspace := findWorkspace(run)
		if block := findBlock(workspace); block != nil {
			blockName = block.Name
		}
		browserUrl := ""
		if workspace != nil {
			browserUrl = fmt.Sprintf(" Logs: %s", app_urls.GetRun(cfg, *workspace, run))
		}
		fmt.Fprintf(os.Stdout, "created run to %s %s and dependencies in %q environment. %s\n", action, blockName, input.Env.Name, browserUrl)
	}
	return nil
}
