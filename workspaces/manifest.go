package workspaces

import (
	"os"
	"path/filepath"

	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/yaml.v3"
)

type Manifest struct {
	OrgName string `json:"orgName" yaml:"org_name"`

	StackId   int64  `json:"stackId" yaml:"stack_id"`
	StackName string `json:"stackName" yaml:"stack_name"`

	BlockId   int64  `json:"blockId" yaml:"block_id"`
	BlockName string `json:"blockName" yaml:"block_name"`
	BlockRef  string `json:"blockRef" yaml:"block_ref"`

	EnvId   int64  `json:"envId" yaml:"env_id"`
	EnvName string `json:"envName" yaml:"env_name"`

	WorkspaceUid string `json:"workspaceUid" yaml:"workspace_uid"`

	// ClassificationLevel is the workspace's data-classification sensitivity slug
	// (e.g. "customer-content"); empty when unclassified. Only the level is
	// threaded (Q6 — level-only tagging); the provider computes the composite
	// cloud-tag value from this slug.
	ClassificationLevel string `json:"classificationLevel,omitempty" yaml:"classification_level,omitempty"`

	// CapabilityId
	// Deprecated
	CapabilityId int64 `json:"capabilityId,omitempty" yaml:"capability_id,omitempty"`

	CapabilityName string `json:"capabilityName,omitempty" yaml:"capability_name,omitempty"`

	Connections ManifestConnections `json:"connections" yaml:"connections"`

	Capabilities ManifestCapabilities `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
}

type ManifestCapabilities map[string]ManifestCapability

type ManifestCapability struct {
	Connections ManifestConnections `json:"connections" yaml:"connections"`
}

func (m Manifest) WriteToFile(filename string) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := yaml.NewEncoder(file)
	defer encoder.Close()
	return encoder.Encode(m)
}

type ManifestConnections map[string]ManifestConnectionTarget

type ManifestConnectionTarget struct {
	StackId   int64  `json:"stackId" yaml:"stack_id"`
	BlockId   int64  `json:"blockId" yaml:"block_id"`
	BlockName string `json:"blockName" yaml:"block_name"`
	EnvId     *int64 `json:"envId,omitempty" yaml:"env_id,omitempty"`
}

// ManifestConnectionsFrom converts resolved workspace connections into manifest targets
// Connections without an effective target are skipped; the CLI surveys those from the user separately
// terraform-provider-ns consults these local targets before falling back to the workspace's run config,
// so every resolved connection must be written for local plans to see the selected configuration
func ManifestConnectionsFrom(conns types.Connections) ManifestConnections {
	result := ManifestConnections{}
	for name, conn := range conns {
		if conn.EffectiveTarget == nil || conn.EffectiveTarget.BlockId < 1 {
			continue
		}
		result[name] = ManifestConnectionTarget{
			StackId:   conn.EffectiveTarget.StackId,
			BlockId:   conn.EffectiveTarget.BlockId,
			BlockName: conn.EffectiveTarget.BlockName,
			EnvId:     conn.EffectiveTarget.EnvId,
		}
	}
	return result
}
