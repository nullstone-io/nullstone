package runs

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func Create(cfg api.Config, workspace types.Workspace, runConfig *types.RunConfig, isApproved *bool, isDestroy bool) (*types.Run, error) {
	input := types.CreateRunInput{
		IsDestroy:         isDestroy,
		IsApproved:        isApproved,
		Source:            runConfig.Source,
		SourceVersion:     runConfig.SourceVersion,
		Variables:         runConfig.Variables,
		EnvVariables:      runConfig.EnvVariables,
		Connections:       runConfig.Connections,
		Capabilities:      runConfig.Capabilities,
		Providers:         runConfig.Providers,
		DependencyConfigs: runConfig.DependencyConfigs,
	}

	client := api.Client{Config: cfg}
	newRun, err := client.Runs().Create(workspace.StackId, workspace.Uid, input)
	if err != nil {
		return nil, fmt.Errorf("error creating run: %w", err)
	}
	return newRun, nil
}
