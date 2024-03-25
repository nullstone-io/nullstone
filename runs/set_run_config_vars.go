package runs

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"strings"
)

// SetConfigVars takes the input vars and stores them as workspace_changes via the API
// TODO: once the api supports it, return any flags that aren't valid for the module version and were skipped
func SetConfigVars(ctx context.Context, cfg api.Config, workspace types.Workspace, varFlags []string) error {
	var variables []types.VariableInput
	for _, varFlag := range varFlags {
		tokens := strings.SplitN(varFlag, "=", 2)
		if len(tokens) < 2 {
			// We skip any variables that don't have an `=` sign
			continue
		}
		name, value := tokens[0], tokens[1]
		variables = append(variables, types.VariableInput{Key: name, Value: value})
	}

	client := api.Client{Config: cfg}
	err := client.WorkspaceVariables().Update(ctx, workspace.StackId, workspace.BlockId, workspace.EnvId, variables)
	if err != nil {
		return fmt.Errorf("failed to update workspace variables: %w", err)
	}

	return nil
}
