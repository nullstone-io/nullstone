package runs

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"strings"
)

// SetConfigVars takes the input vars and stores them as workspace_changes via the API
// TODO: once the api supports it, return any flags that aren't valid for the module version and were skipped
func SetConfigVars(cfg api.Config, workspace types.Workspace, varFlags []string) ([]types.WorkspaceChange, error) {
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
	changes, err := client.WorkspaceVariables().Update(workspace.StackId, workspace.Uid, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace variables: %w", err)
	}

	return changes, nil
}
