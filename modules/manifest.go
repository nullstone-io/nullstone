package modules

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func ManifestFromFile(filename string) (*types.ModuleManifest, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("module manifest file %q does not exist", filename)
		}
	}
	defer file.Close()

	manifest := types.ModuleManifest{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("error decoding module manifest: %w", err)
	}
	return &manifest, nil
}

func WriteManifestToFile(m types.ModuleManifest, filename string) error {
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
