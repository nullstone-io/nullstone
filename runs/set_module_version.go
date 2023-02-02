package runs

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func SetModuleVersion(cfg api.Config, workspace types.Workspace, input types.WorkspaceModuleInput) error {
	client := api.Client{Config: cfg}
	err := client.WorkspaceModule().Update(workspace.StackId, workspace.BlockId, workspace.EnvId, input)
	if err != nil {
		return fmt.Errorf("failed to update workspace variables: %w", err)
	}

	return nil
}
