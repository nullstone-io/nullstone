package workspaces

import (
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// GetRunConfig loads the effective run config for a workspace
// This does the following:
//   1. Pull the latest run config for the workspace
//   2. Scan module in local file system for `ns_connection` that have not been added to run config
func GetRunConfig(cfg api.Config, workspace Manifest) (types.RunConfig, error) {
	client := api.Client{Config: cfg}
	uid, _ := uuid.Parse(workspace.WorkspaceUid)
	runConfig, err := client.RunConfigs().GetLatest(workspace.StackId, uid)
	if err != nil {
		return types.RunConfig{}, err
	} else if runConfig == nil {
		runConfig = &types.RunConfig{
			WorkspaceUid:  uid,
			Source:        "",
			SourceVersion: "",
			Variables:     types.Variables{},
			Connections:   types.Connections{},
			Capabilities:  types.CapabilityConfigs{},
			Providers:     types.Providers{},
			Targets:       types.RunTargets{},
			Dependencies:  types.Dependencies{},
		}
	}

	// Scan module in local file system
	localManifest, err := ScanLocal(".")
	if err != nil {
		return *runConfig, fmt.Errorf("could not scan local module: %w", err)
	}

	// Look for new connections locally that aren't present in the workspace's run config
	for name, local := range localManifest.Connections {
		_, ok := runConfig.Connections[name]
		if !ok {
			// Connection exists in local scan, but not in run config
			// Let's add the definition with an empty target
			runConfig.Connections[name] = types.Connection{
				Connection: local,
				Target:     "",
				Reference:  nil,
				Unused:     false,
			}
		}
	}
	return *runConfig, nil
}
