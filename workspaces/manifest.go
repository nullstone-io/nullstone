package workspaces

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
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

	CapabilityId int64 `json:"capabilityId,omitempty" yaml:"capability_id,omitempty"`

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
