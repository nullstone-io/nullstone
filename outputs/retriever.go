package outputs

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"reflect"
)

func Retrieve[T any](nsConfig api.Config, workspace *types.Workspace) (T, error) {
	var t T
	r := Retriever{NsConfig: nsConfig}
	if err := r.Retrieve(workspace, &t); err != nil {
		return t, err
	}
	return t, nil
}

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

	nsClient := api.Client{Config: r.NsConfig}
	workspaceOutputs, err := nsClient.WorkspaceOutputs().GetLatest(workspace.StackId, workspace.BlockId, workspace.EnvId)
	if workspaceOutputs == nil {
		return fmt.Errorf("this workspace has not yet been successfully launched, no outputs exist yet")
	}
	if err != nil {
		wt := types.WorkspaceTarget{
			StackId: workspace.StackId,
			BlockId: workspace.BlockId,
			EnvId:   workspace.EnvId,
		}
		return fmt.Errorf("unable to fetch the outputs for %s/%s", workspace.OrgName, wt.Id())
	}

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

			connWorkspace, err := r.GetConnectionWorkspace(workspace, field.ConnectionName, field.ConnectionType, field.ConnectionContract)
			if err != nil {
				return fmt.Errorf("error finding connection workspace (name=%s, type=%s, contract=%s): %w", field.ConnectionName, field.ConnectionType, field.ConnectionContract, err)
			}
			if connWorkspace == nil {
				if field.Optional {
					continue
				}
				return ErrMissingRequiredConnection{
					ConnectionName:     field.ConnectionName,
					ConnectionType:     field.ConnectionType,
					ConnectionContract: field.ConnectionContract,
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
			if err := field.SafeSet(obj, *workspaceOutputs); err != nil {
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
func (r *Retriever) GetConnectionWorkspace(source *types.Workspace, connectionName, connectionType, connectionContract string) (*types.Workspace, error) {
	conn, err := findConnection(source, connectionName, connectionType, connectionContract)
	if err != nil {
		return nil, err
	} else if conn == nil || conn.Reference == nil {
		return nil, nil
	}

	sourceTarget := types.WorkspaceTarget{
		StackId: source.StackId,
		BlockId: source.BlockId,
		EnvId:   source.EnvId,
	}
	destTarget := sourceTarget.FindRelativeConnection(*conn.Reference)

	nsClient := api.Client{Config: r.NsConfig}
	return nsClient.Workspaces().Get(destTarget.StackId, destTarget.BlockId, destTarget.EnvId)
}

func findConnection(source *types.Workspace, connectionName, connectionType, connectionContract string) (*types.Connection, error) {
	if source.LastFinishedRun == nil || source.LastFinishedRun.Config == nil {
		return nil, fmt.Errorf("cannot find connections for app")
	}
	hasType := connectionType != ""
	hasContract := connectionContract != ""
	if connectionName == "" && (!hasType && !hasContract) {
		return nil, fmt.Errorf("cannot find connection if name or type/contract is not specified")
	}
	var desiredContract types.ModuleContractName
	if hasContract {
		var err error
		if desiredContract, err = types.ParseModuleContractName(connectionContract); err != nil {
			return nil, fmt.Errorf("invalid connection contract %q: %w", connectionContract, err)
		}
	}

	for name, connection := range source.LastFinishedRun.Config.Connections {
		curConnContract, err := types.ParseModuleContractName(connection.Contract)
		if err != nil {
			// We are skipping connections with bad contracts in the current run config
			continue
		}
		if hasContract && !desiredContract.Match(curConnContract) {
			continue
		}
		if hasType && connectionType != connection.Type {
			continue
		}
		if connectionName != "" && connectionName != name {
			continue
		}
		return &connection, nil
	}
	return nil, nil
}
