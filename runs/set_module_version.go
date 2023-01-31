package runs

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func SetModuleVersion(cfg api.Config, workspace types.Workspace, version string) (*types.WorkspaceChangeset, error) {
	client := api.Client{Config: cfg}
	changes, err := client.WorkspaceModuleVersion().Update(workspace.StackId, workspace.BlockId, workspace.EnvId, version)
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace variables: %w", err)
	}

	return changes, nil
}
