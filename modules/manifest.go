package modules

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Manifest struct {
	OrgName       string   `yaml:"org_name"`
	Name          string   `yaml:"name"`
	FriendlyName  string   `yaml:"friendly_name"`
	Description   string   `yaml:"description"`
	Category      string   `yaml:"category"`
	Type          string   `yaml:"type"`
	Layer         string   `yaml:"layer"`
	IsPublic      bool     `yaml:"is_public"`
	ProviderTypes []string `yaml:"provider_types"`
}

func ManifestFromFile(filename string) (*Manifest, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("module manifest file %q does not exist", filename)
		}
	}
	defer file.Close()

	manifest := Manifest{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("error decoding module manifest: %w", err)
	}
	return &manifest, nil
}

func (m Manifest) WriteManifestToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := yaml.NewEncoder(file)
	defer encoder.Close()
	return encoder.Encode(m)
}
