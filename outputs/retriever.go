package outputs

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"reflect"
)

type Retriever struct {
	NsConfig api.Config
}

// Retrieve is capable of retrieving all outputs for a given workspace
// To properly use, the input obj must be a pointer to a struct that contains fields that map to outputs
// Struct tags on each field within the struct define how to read the outputs from nullstone APIs
// See Field for more details
func (r *Retriever) Retrieve(workspace *types.Workspace, obj interface{}) error {
	objType := reflect.TypeOf(obj)
	if objType.Kind() != reflect.Ptr {
		return fmt.Errorf("input object must be a pointer")
	}
	if objType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("input object must be a pointer to a struct")
	}

	if workspace.LastSuccessfulRun == nil || workspace.LastSuccessfulRun.Apply == nil {
		wt := types.WorkspaceTarget{
			OrgName:   workspace.OrgName,
			StackName: workspace.StackName,
			BlockName: workspace.BlockName,
			EnvName:   workspace.EnvName,
		}
		return fmt.Errorf("cannot find outputs for %s", wt.Id())
	}
	workspaceOutputs := workspace.LastSuccessfulRun.Apply.Outputs

	fields := GetFields(reflect.TypeOf(obj).Elem())
	for _, field := range fields {
		fieldType := field.Field.Type

		if field.Name == "" {
			// `ns:",..."` refers to a connection, this field must be a struct type
			//we're going to run retrieve into this field
			if err := CheckValidConnectionField(obj, fieldType); err != nil {
				return err
			}
			target := field.InitializeConnectionValue(obj)

			connWorkspace, err := r.GetConnectionWorkspace(workspace, field.ConnectionName, field.ConnectionType)
			if err != nil {
				return fmt.Errorf("error finding connection workspace (name=%s, type=%s): %w", field.ConnectionName, field.ConnectionType, err)
			}
			if connWorkspace == nil {
				if field.Optional {
					continue
				}
				return ErrMissingRequiredConnection{
					ConnectionName: field.ConnectionName,
					ConnectionType: field.ConnectionType,
				}
			}
			if err := r.Retrieve(connWorkspace, target); err != nil {
				return err
			}
		} else {
			// `ns:"xyz"` refers to an output named `xyz` in the current workspace outputs
			// we're going to extract the value into this field
			if err := CheckValidField(obj, fieldType); err != nil {
				return err
			}
			if err := field.SafeSet(obj, workspaceOutputs); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetConnectionWorkspace gets the workspace from nullstone through a connection from the source workspace
// This will search through connections matching on connectionName and connectionType
// Specify "" to ignore filtering for that field
// One of either connectionName or connectionType must be specified
func (r *Retriever) GetConnectionWorkspace(source *types.Workspace, connectionName, connectionType string) (*types.Workspace, error) {
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

	nsClient := api.Client{Config: r.NsConfig}
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
