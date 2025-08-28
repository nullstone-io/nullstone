package iac

import (
	"context"
	"fmt"
	"io"

	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/iac"
	iacEvents "github.com/nullstone-io/iac/events"
	"github.com/nullstone-io/iac/workspace"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

const (
	indentStep = "    "
)

func Test(ctx context.Context, cfg api.Config, w io.Writer, stack types.Stack, env types.Environment, pmr iac.ConfigFiles) error {
	if err := testWorkspaces(ctx, cfg, w, stack, env, pmr); err != nil {
		return err
	}
	if err := testEvents(ctx, cfg, w, stack, env, pmr); err != nil {
		return err
	}
	return nil
}

func testWorkspaces(ctx context.Context, cfg api.Config, w io.Writer, stack types.Stack, env types.Environment, pmr iac.ConfigFiles) error {
	blockNames := pmr.BlockNames(env)
	apiClient := &api.Client{Config: cfg}
	allBlocks, err := apiClient.Blocks().List(ctx, stack.Id, false)
	if err != nil {
		return fmt.Errorf("error retrieving blocks: %w", err)
	}

	blocks := make(types.Blocks, 0)
	for _, block := range allBlocks {
		if _, ok := blockNames[block.Name]; ok {
			blocks = append(blocks, block)
		}
	}

	hasError := false
	plural := "s"
	if len(blocks) == 1 {
		plural = ""
	}
	colorstring.Fprintf(w, "[bold]Detecting changes for %d block%s in %s/%s...[reset]\n", len(blocks), plural, stack.Name, env.Name)
	for _, block := range blocks {
		if err := testWorkspace(ctx, apiClient, w, stack, block, env, pmr); err != nil {
			colorstring.Fprintf(w, "[red]An error occurred: %s[reset]\n", err)
			hasError = true
		}
	}

	if hasError {
		return fmt.Errorf("An error occurred diffing block%s.", plural)
	}
	return nil
}

func testWorkspace(ctx context.Context, apiClient *api.Client, w io.Writer, stack types.Stack, block types.Block, env types.Environment, pmr iac.ConfigFiles) error {
	effective, err := apiClient.WorkspaceConfigs().GetEffective(ctx, stack.Id, block.Id, env.Id)
	if err != nil {
		return fmt.Errorf("error retrieving workspace: %w", err)
	} else if effective == nil {
		return nil
	}

	fillWorkspaceConfigMissingEnv(effective, env)

	updated, err := effective.Clone()
	if err != nil {
		return fmt.Errorf("error cloning workspace: %w", err)
	}

	updater := workspace.ConfigUpdater{
		Config: &updated,
		TemplateVars: workspace.TemplateVars{
			OrgName:   stack.OrgName,
			StackName: stack.Name,
			EnvName:   env.Name,
			EnvIsProd: env.IsProd,
		},
	}
	if err := iac.ApplyChangesTo(pmr, block, env, updater); err != nil {
		return fmt.Errorf("error applying changes: %w", err)
	}

	changes := workspace.DiffWorkspaceConfig(*effective, updated)
	emitWorkspaceChanges(w, block, changes)
	return nil
}

func testEvents(ctx context.Context, cfg api.Config, w io.Writer, stack types.Stack, env types.Environment, pmr iac.ConfigFiles) error {
	apiClient := api.Client{Config: cfg}
	existingEvents, err := apiClient.EnvEvents().List(ctx, stack.Id, env.Id)
	if err != nil {
		return fmt.Errorf("error looking up existing events: %w", err)
	}

	current := map[string]types.EnvEvent{}
	for _, cur := range existingEvents {
		if cur.OwningRepoUrl == pmr.RepoUrl {
			current[cur.Name] = cur
		}
	}

	colorstring.Fprintf(w, "[bold]Detecting changes for events in %s/%s...[reset]\n", stack.Name, env.Name)
	desired := iacEvents.Get(pmr, env)
	changes := iacEvents.Diff(current, desired, pmr.RepoUrl)
	emitEventChanges(w, changes)
	return nil
}

func fillWorkspaceConfigMissingEnv(c *types.WorkspaceConfig, env types.Environment) {
	envId := env.Id
	fillRef := func(conn types.Connection) bool {
		if conn.EffectiveTarget == nil {
			return false
		}
		filled := false
		if conn.EffectiveTarget.StackId == env.StackId {
			if conn.EffectiveTarget.EnvId == nil {
				conn.EffectiveTarget.EnvId = &envId
				filled = true
			}
			if conn.EffectiveTarget.EnvName == "" {
				conn.EffectiveTarget.EnvName = env.Name
				filled = true
			}
		}
		return filled
	}
	fillConns := func(conns types.Connections) bool {
		filled := false
		for name, conn := range conns {
			if fillRef(conn) {
				conns[name] = conn
				filled = true
			}
		}
		return filled
	}

	fillConns(c.Connections)
	for i, capability := range c.Capabilities {
		if fillConns(capability.Connections) {
			c.Capabilities[i] = capability
		}
	}
}
