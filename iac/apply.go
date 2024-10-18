package iac

import (
	"context"
	"fmt"
	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/iac"
	"github.com/nullstone-io/iac/workspace"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"io"
)

func Apply(ctx context.Context, cfg api.Config, curDir string, w io.Writer, stack types.Stack, env types.Environment, pmr iac.ParseMapResult) error {
	blockNames := pmr.BlockNames(env)
	apiClient := &api.Client{Config: cfg}
	allBlocks, err := apiClient.Blocks().List(ctx, stack.Id)
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
	colorstring.Fprintf(w, "[bold]Detecting changes for %d blocks in %s/%s...[reset]\n", len(blocks), stack.Name, env.Name)
	for _, block := range blocks {
		if err := applyWorkspace(ctx, apiClient, w, stack, block, env, pmr); err != nil {
			colorstring.Fprintf(w, "[red]An error occurred: %s[reset]\n", err)
			hasError = true
		}
	}

	if hasError {
		return fmt.Errorf("An error occurred diffing blocks.")
	}
	return nil
}

func applyWorkspace(ctx context.Context, apiClient *api.Client, w io.Writer, stack types.Stack, block types.Block, env types.Environment, pmr iac.ParseMapResult) error {
	effective, err := apiClient.WorkspaceConfigs().GetEffective(ctx, stack.Id, block.Id, env.Id)
	if err != nil {
		return fmt.Errorf("error retrieving workspace: %w", err)
	} else if effective == nil {
		return nil
	}

	updated, err := effective.Clone()
	if err != nil {
		return fmt.Errorf("error cloning workspace: %w", err)
	}

	if err := iac.ApplyChangesTo(pmr, block, env, workspace.ConfigUpdater{Config: &updated}); err != nil {
		return fmt.Errorf("error applying changes: %w", err)
	}

	changes := workspace.DiffWorkspaceConfig(*effective, updated)
	if len(changes) == 0 {
		return nil
	}

	s := "s"
	if len(changes) == 1 {
		s = ""
	}
	colorstring.Fprintf(w, "    [bold]%s[reset] => %d change%s\n", block.Name, len(changes), s)
	for _, change := range changes {
		identifier := fmt.Sprintf(".%s", change.Identifier)
		if identifier == types.ChangeIdentifierModuleVersion {
			identifier = ""
		} else if change.ChangeType == types.ChangeTypeCapability {
			if cur, ok := change.Current.(types.CapabilityConfig); ok {
				identifier = fmt.Sprintf("[%s]", cur.Source)
			}
			if desired, ok := change.Desired.(types.CapabilityConfig); identifier == "" && ok {
				identifier = fmt.Sprintf("[%s]", desired.Source)
			}
		}
		switch change.Action {
		case types.ChangeActionAdd:
			colorstring.Fprintf(w, "        [green]+ %s%s[reset]\n", change.ChangeType, identifier)
		case types.ChangeActionDelete:
			colorstring.Fprintf(w, "        [red]- %s%s[reset]\n", change.ChangeType, identifier)
		case types.ChangeActionUpdate:
			colorstring.Fprintf(w, "        [yellow]~ %s%s[reset]\n", change.ChangeType, identifier)
		}
	}

	return nil
}
