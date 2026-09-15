package workspaces

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// ConfigSource selects which workspace configuration to pull from Nullstone
type ConfigSource string

const (
	// ConfigSourceEffective is the latest configuration including unapplied changes queued in the UI
	// This is the configuration that will be used the next time the workspace is applied
	ConfigSourceEffective ConfigSource = "effective"
	// ConfigSourceLatest is the latest configuration that was applied (its run may still be in progress)
	ConfigSourceLatest ConfigSource = "latest"
	// ConfigSourceCurrent is the configuration from the last finished run
	ConfigSourceCurrent ConfigSource = "current"
)

var ConfigSources = []ConfigSource{ConfigSourceEffective, ConfigSourceLatest, ConfigSourceCurrent}

func ConfigSourceNames() []string {
	names := make([]string, 0, len(ConfigSources))
	for _, cs := range ConfigSources {
		names = append(names, string(cs))
	}
	return names
}

// ParseConfigSource validates a user-provided config source
// An empty value defaults to ConfigSourceEffective
func ParseConfigSource(raw string) (ConfigSource, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ConfigSourceEffective, nil
	}
	for _, cs := range ConfigSources {
		if raw == string(cs) {
			return cs, nil
		}
	}
	return "", fmt.Errorf("invalid --config %q: must be one of %s", raw, strings.Join(ConfigSourceNames(), ", "))
}

// GetWorkspaceConfig loads the workspace config for a workspace from the requested source
// This does the following:
//  1. Pull the workspace config from Nullstone (see ConfigSource for what each source represents)
//  2. Scan module in local file system for `ns_connection` that have not been added to the workspace config
func GetWorkspaceConfig(ctx context.Context, cfg api.Config, workspace Manifest, source ConfigSource) (types.WorkspaceConfig, error) {
	client := api.Client{Config: cfg}
	var config *types.WorkspaceConfig
	var err error
	switch source {
	case ConfigSourceEffective:
		config, err = client.WorkspaceConfigs().GetEffective(ctx, workspace.StackId, workspace.BlockId, workspace.EnvId)
	case ConfigSourceLatest:
		config, err = client.WorkspaceConfigs().GetLatest(ctx, workspace.StackId, workspace.BlockId, workspace.EnvId)
	case ConfigSourceCurrent:
		config, err = client.WorkspaceConfigs().GetCurrent(ctx, workspace.StackId, workspace.BlockId, workspace.EnvId)
	default:
		return types.WorkspaceConfig{}, fmt.Errorf("unknown workspace config source %q", source)
	}
	if err != nil {
		return types.WorkspaceConfig{}, err
	} else if config == nil {
		// The workspace has no configuration for this source yet (e.g. `current` before the first run finishes)
		config = &types.WorkspaceConfig{
			Source:        "",
			SourceVersion: "",
			Variables:     types.Variables{},
			Connections:   types.Connections{},
			Capabilities:  types.CapabilityConfigs{},
			Providers:     types.Providers{},
			Dependencies:  types.Dependencies{},
		}
	}
	if config.Connections == nil {
		config.Connections = types.Connections{}
	}

	// Scan module in local file system
	localManifest, err := ScanLocal(".")
	if err != nil {
		return *config, fmt.Errorf("could not scan local module: %w", err)
	}

	// Look for new connections locally that aren't present in the workspace config
	for name, local := range localManifest.Connections {
		_, ok := config.Connections[name]
		if !ok {
			// Connection exists in local scan, but not in workspace config
			// Let's add the definition with an empty target
			config.Connections[name] = types.Connection{
				Connection:      local,
				DesiredTarget:   nil,
				EffectiveTarget: nil,
				Unused:          false,
			}
		}
	}
	return *config, nil
}
