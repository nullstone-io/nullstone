package runs

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func Create(ctx context.Context, cfg api.Config, workspace types.Workspace, isApproved *bool, isDestroy bool) (*types.Run, error) {
	input := types.CreateRunInput{
		IsDestroy:  isDestroy,
		IsApproved: isApproved,
	}

	client := api.Client{Config: cfg}
	newRun, err := client.Runs().Create(ctx, workspace.StackId, workspace.Uid, input)
	if err != nil {
		return nil, fmt.Errorf("error creating run: %w", err)
	}
	return newRun, nil
}
