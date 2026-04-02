package modules

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/yaml.v3"
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

func WriteManifestToLogger(manifest types.ModuleManifest, logger *log.Logger) {
	url := manifest.SourceUrl
	if url == "" {
		url = "<not configured>"
	}
	subcategory := ""
	if manifest.Subcategory != "" {
		subcategory = fmt.Sprintf(":%s", manifest.Subcategory)
	}
	provider := "*"
	if len(manifest.ProviderTypes) > 0 {
		provider = strings.Join(manifest.ProviderTypes, ",")
	}
	subplatform := ""
	if manifest.Subplatform != "" {
		subplatform = fmt.Sprintf(":%s", manifest.Subplatform)
	}

	logger.Println(fmt.Sprintf("Module: %s/%s", manifest.OrgName, manifest.Name))
	logger.Println(fmt.Sprintf("URL: %s", manifest.SourceUrl))
	logger.Println(fmt.Sprintf("Contract: %s%s/%s/%s%s", manifest.Category, subcategory, provider, manifest.Platform, subplatform))
	logger.Println(fmt.Sprintf("Tool: %s", manifest.ToolName))
}
