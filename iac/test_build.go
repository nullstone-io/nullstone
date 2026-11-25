package iac

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/iac"
	"github.com/nullstone-io/iac/config"
	"github.com/nullstone-io/iac/workspace"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// testBuildConfig reports the build config settings an IaC sync would write for an application workspace.
//
// A build config is not part of the workspace config: a sync writes it straight to the workspace and it
// applies to the next build, so these are reported apart from the workspace changes.
// IaC owns a build config only where a file declares `build:` for the app. An app with no `build:` in any
// file is left alone: if IaC wrote its build config earlier, the settings stay as they are.
//
// Settings that the builder or the package mode of the app's module would not use are an error.
// Validation reports them where the files name the builder and the module; here they are checked
// against the workspace, which knows both.
func testBuildConfig(ctx context.Context, apiClient *api.Client, w io.Writer, stack types.Stack, block types.Block, env types.Environment, pmr iac.ConfigFiles) error {
	if block.Type != string(types.BlockTypeApplication) {
		return nil
	}
	current, err := apiClient.WorkspaceBuildConfigs().Get(ctx, stack.Id, block.Id, env.Id)
	if err != nil {
		return fmt.Errorf("error retrieving build config: %w", err)
	} else if current == nil {
		return nil
	}
	if !declaresBuild(pmr, block, env) {
		return nil
	}

	template, err := apiClient.WorkspaceTemplates().Get(ctx, stack.Id, block.Id)
	if err != nil {
		return fmt.Errorf("error retrieving workspace template: %w", err)
	}
	module, err := findBuildModule(ctx, apiClient.Config, block, template, pmr)
	if err != nil {
		return err
	}
	desired, err := desiredBuildConfig(pmr, block, env, template, module)
	if err != nil {
		return err
	}

	emitBuildChanges(w, block, workspace.DiffBuildConfig(current.Config, desired))
	return nil
}

// declaresBuild reports whether the files declare `build:` for block in env
func declaresBuild(pmr iac.ConfigFiles, block types.Block, env types.Environment) bool {
	probe := types.BuildConfig{}
	updater := &workspace.BuildConfigUpdater{Config: &probe}
	if err := iac.ApplyBuildChangesTo(pmr, block, env, updater); err != nil {
		return false
	}
	return updater.Source != nil
}

// desiredBuildConfig resolves the build config the files declare for block in env.
// It starts from what the app falls back to: its block template, else the module default.
// The package mode always comes from the module.
func desiredBuildConfig(pmr iac.ConfigFiles, block types.Block, env types.Environment, template *types.WorkspaceTemplate, module *types.Module) (types.BuildConfig, error) {
	moduleDefault := types.DefaultBuildConfig(module)
	fallback := moduleDefault
	if template != nil && template.Config.Build != nil {
		fallback = *template.Config.Build
	}

	fit := config.BuildTarget{Builder: fallback.Builder, PackageMode: moduleDefault.PackageMode, DockerContext: fallback.DockerContext}
	if errs := iac.ValidateBuild(pmr, block, env, fit); len(errs) > 0 {
		messages := make([]string, 0, len(errs))
		for _, cur := range errs {
			messages = append(messages, fmt.Sprintf("%s => %s", cur.ObjectPathContext.Context(), cur.ErrorMessage))
		}
		return fallback, fmt.Errorf("invalid build config: %s", strings.Join(messages, "; "))
	}

	desired := workspace.CloneBuildConfig(fallback)
	updater := &workspace.BuildConfigUpdater{Config: &desired, ModuleDefault: moduleDefault}
	if err := iac.ApplyBuildChangesTo(pmr, block, env, updater); err != nil {
		return desired, fmt.Errorf("error applying build config: %w", err)
	}
	desired.PackageMode = moduleDefault.PackageMode
	return desired, nil
}

// findBuildModule finds the module that decides the default build config of an app.
// The module that config.yml declares wins, since a sync sets it before it resolves the build config.
func findBuildModule(ctx context.Context, cfg api.Config, block types.Block, template *types.WorkspaceTemplate, pmr iac.ConfigFiles) (*types.Module, error) {
	moduleSource := block.ModuleSource
	if template != nil && template.Config.Module != "" {
		moduleSource = template.Config.Module
	}
	if pmr.Config != nil {
		if app, ok := pmr.Config.Applications[block.Name]; ok {
			if app.Module != nil {
				return app.Module, nil
			}
			if app.ModuleSource != "" {
				moduleSource = app.ModuleSource
			}
		}
	}
	if moduleSource == "" {
		return nil, nil
	}
	module, err := find.Module(ctx, cfg, moduleSource)
	if err != nil {
		return nil, fmt.Errorf("error retrieving module (%s): %w", moduleSource, err)
	}
	return module, nil
}

func emitBuildChanges(w io.Writer, block types.Block, changes []types.BuildConfigChange) {
	if len(changes) == 0 {
		return
	}

	s := "s"
	if len(changes) == 1 {
		s = ""
	}
	indent := indentStep
	colorstring.Fprintf(w, "%s[bold]%s[reset] => %d build config change%s (applied to the next build)\n", indent, block.Name, len(changes), s)
	indent += indentStep
	for _, change := range changes {
		colorstring.Fprintf(w, "%s[yellow]~ build.%s[reset]: [red]%s[reset] => [green]%s[reset]\n",
			indent, change.Key, buildValToString(change.Current), buildValToString(change.Desired))
	}
}

func buildValToString(val any) string {
	switch v := val.(type) {
	case string:
		if v == "" {
			return "(empty)"
		}
		return v
	case bool:
		return strconv.FormatBool(v)
	case map[string]string:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(v))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s: %s", k, v[k]))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	}
	return variableValToString(val)
}
