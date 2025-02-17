package workspaces

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0/artifacts"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

var (
	capabilitiesTemplateFuncs template.FuncMap
)

func init() {
	toJson := func(v interface{}) string {
		if v == nil {
			return "null"
		}
		rawJson, _ := json.Marshal(v)
		return string(rawJson)
	}
	// to_json_string is intended to emit the json as a string
	// This is helpful when wrapping in terraform with `jsondecode(...)`
	toJsonString := func(v interface{}) string {
		return toJson(toJson(v))
	}

	capabilitiesTemplateFuncs = template.FuncMap{
		"to_json":        toJson,
		"to_json_string": toJsonString,
	}
}

type CapabilitiesGenerator struct {
	RegistryAddress  string
	Manifest         Manifest
	TemplateFilename string
	TargetFilename   string
}

func (g CapabilitiesGenerator) ShouldGenerate() bool {
	_, err := os.Lstat(g.TemplateFilename)
	return err == nil || !os.IsNotExist(err)
}

func (g CapabilitiesGenerator) Generate(runConfig types.RunConfig) error {
	capabilities, err := g.transformCapabilities(runConfig)
	if err != nil {
		return fmt.Errorf("error retrieving current configuration of capabilities: %w", err)
	}

	rawTemplateContent, err := ioutil.ReadFile(g.TemplateFilename)
	if err != nil {
		return fmt.Errorf("error reading capabilities template file: %w", err)
	}

	content := bytes.NewBufferString("")
	tmpl, err := template.New("capabilities").Funcs(capabilitiesTemplateFuncs).Parse(string(rawTemplateContent))
	if err != nil {
		return fmt.Errorf("error parsing capabilities template: %w", err)
	}

	if err := tmpl.Execute(content, capabilities); err != nil {
		return fmt.Errorf("error executing capabilities template: %w", err)
	}

	if err := ioutil.WriteFile(g.TargetFilename, content.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing %q: %s", g.TargetFilename, err)
	}
	return nil
}

func (g CapabilitiesGenerator) transformCapabilities(runConfig types.RunConfig) (types.CapabilityConfigs, error) {
	// Terraform assumes that module source has a host of `registry.terraform.io` if not specified
	// We are going to override that behavior to presume `api.nullstone.io` instead
	capabilities := runConfig.Capabilities
	for i, capability := range capabilities {
		if ms, err := artifacts.ParseSource(capability.Source); err == nil {
			if ms.Host == "" {
				// Set the module source host to api.nullstone.io without the URI scheme
				ms.Host = strings.TrimPrefix(strings.TrimPrefix(g.RegistryAddress, "https://"), "http://")
				capability.Source = ms.String()
				capabilities[i] = capability
			}
		}
	}
	return capabilities, nil
}
