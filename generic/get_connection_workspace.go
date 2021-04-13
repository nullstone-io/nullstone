package generic

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// GetConnectionWorkspace gets the workspace from nullstone through a connection from the source workspace
// This will search through connections matching on connectionName and connectionType
// Specify "" for either to ignore filtering
// One of either connectionName or connectionType must be specified
func GetConnectionWorkspace(nsConfig api.Config, source *types.Workspace, connectionName, connectionType string) (*types.Workspace, error) {
	conn, err := findConnection(source, connectionName, connectionType)
	if err != nil {
		return nil, err
	} else if conn == nil {
		return nil, nil
	}

	sourceTarget := types.WorkspaceTarget{
		OrgName:   source.OrgName,
		StackName: source.StackName,
		BlockName: source.BlockName,
		EnvName:   source.EnvName,
	}
	destTarget := sourceTarget.FindRelativeConnection(conn.Target)

	nsClient := api.Client{Config: nsConfig}
	return nsClient.Workspaces().Get(destTarget.StackName, destTarget.BlockName, destTarget.EnvName)
}

func findConnection(source *types.Workspace, connectionName, connectionType string) (*types.Connection, error) {
	if source.LastSuccessfulRun == nil || source.LastSuccessfulRun.Config == nil {
		return nil, fmt.Errorf("cannot find connections for app")
	}
	if connectionName == "" && connectionType == "" {
		return nil, fmt.Errorf("cannot find connection if name or type is not specified")
	}
	for name, connection := range source.LastSuccessfulRun.Config.Connections {
		if connectionType != "" && connectionType != connection.Type {
			continue
		}
		if connectionName != "" && connectionName != name {
			continue
		}
		return &connection, nil
	}
	return nil, nil
}
