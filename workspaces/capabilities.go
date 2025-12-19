package workspaces

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"text/template"

	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/artifacts"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
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
	ApiConfig        api.Config
}

func (g CapabilitiesGenerator) ShouldGenerate() bool {
	_, err := os.Lstat(g.TemplateFilename)
	return err == nil || !os.IsNotExist(err)
}

func (g CapabilitiesGenerator) Generate(runConfig types.RunConfig) error {
	capabilities := runConfig.Capabilities
	var err error
	if capabilities, err = g.backfillMeta(capabilities); err != nil {
		return fmt.Errorf("error filling capability meta: %w", err)
	}
	if capabilities, err = g.transformCapabilities(capabilities); err != nil {
		return fmt.Errorf("error retrieving current configuration of capabilities: %w", err)
	}

	rawTemplateContent, err := os.ReadFile(g.TemplateFilename)
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

	if err := os.WriteFile(g.TargetFilename, content.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing %q: %s", g.TargetFilename, err)
	}
	return nil
}

func (g CapabilitiesGenerator) transformCapabilities(capabilities types.CapabilityConfigs) (types.CapabilityConfigs, error) {
	// Terraform assumes that module source has a host of `registry.terraform.io` if not specified
	// We are going to override that behavior to presume `api.nullstone.io` instead
	result := make(types.CapabilityConfigs, 0)
	for _, capability := range capabilities {
		if ms, err := artifacts.ParseSource(capability.Source); err == nil {
			if ms.Host == "" {
				// Set the module source host to api.nullstone.io without the URI scheme
				ms.Host = strings.TrimPrefix(strings.TrimPrefix(g.RegistryAddress, "https://"), "http://")
				capability.Source = ms.String()
			}
		}
		result = append(result, capability)
	}
	return result, nil
}

func (g CapabilitiesGenerator) backfillMeta(capabilities types.CapabilityConfigs) (types.CapabilityConfigs, error) {
	errs := make([]error, 0)
	result := make(types.CapabilityConfigs, 0)
	for _, cur := range capabilities {
		meta, err := g.resolveCapabilityMeta(cur)
		if err != nil {
			errs = append(errs, err)
		}
		cur.Meta = meta
		result = append(result, cur)
	}
	return result, errors.Join(errs...)
}

func (g CapabilitiesGenerator) resolveCapabilityMeta(cur types.CapabilityConfig) (*types.CapabilityConfigMeta, error) {
	ctx := context.Background()
	mod, err := find.Module(ctx, g.ApiConfig, cur.Source)
	if err != nil {
		return nil, err
	}
	meta := &types.CapabilityConfigMeta{
		Subcategory: mod.Subcategory,
		Platform:    mod.Platform,
		Subplatform: mod.Subplatform,
	}

	mv, err := find.ModuleVersion(ctx, g.ApiConfig, cur.Source, cur.SourceVersion)
	if err != nil {
		return nil, err
	}
	meta.OutputNames = slices.Collect(maps.Keys(mv.Manifest.Outputs))
	// env + secrets don't show up in the manifest outputs; we should always include them
	meta.OutputNames = append(meta.OutputNames, "env", "secrets")
	if meta.OutputNames == nil {
		meta.OutputNames = make([]string, 0)
	}

	return meta, nil
}
