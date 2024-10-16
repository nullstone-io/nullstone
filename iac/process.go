package iac

import (
	"context"
	"fmt"
	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/iac"
	"github.com/nullstone-io/iac/core"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"io"
	"path/filepath"
)

func Process(ctx context.Context, cfg api.Config, curDir string, w io.Writer, stack types.Stack, env types.Environment, pmr iac.ParseMapResult) error {
	apiClient := &api.Client{Config: cfg}
	colorstring.Fprintf(w, "[bold]Testing Nullstone IaC files against %s/%s environment...[reset]\n", stack.Name, env.Name)
	resolver := core.NewApiResolver(apiClient, stack.Id, env.Id)
	if errs := iac.Resolve(ctx, pmr, resolver); len(errs) > 0 {
		colorstring.Fprintf(w, "[bold]Detected errors when resolving Nullstone IaC files[reset]\n")
		for _, err := range errs {
			relFilename, _ := filepath.Rel(curDir, err.IacContext.Filename)
			colorstring.Fprintf(w, "    [red]✖[reset] (%s) %s => %s\n", relFilename, err.ObjectPathContext.Context(), err.ErrorMessage)
		}
		fmt.Fprintln(w)
		return fmt.Errorf("IaC files are invalid.")
	} else {
		colorstring.Fprintln(w, "    [green]✔[reset] Resolution completed successfully.")
	}

	if errs := iac.Validate(pmr); len(errs) > 0 {
		colorstring.Fprintf(w, "    [bold]Detected errors when validating Nullstone IaC files[reset]\n")
		for _, err := range errs {
			relFilename, _ := filepath.Rel(curDir, err.IacContext.Filename)
			colorstring.Fprintf(w, "        [red]✖[reset] (%s) %s => %s\n", relFilename, err.ObjectPathContext.Context(), err.ErrorMessage)
		}
		fmt.Fprintln(w)
		return fmt.Errorf("IaC files are invalid.")
	} else {
		colorstring.Fprintln(w, "    [green]✔[reset] Validation completed successfully.")
	}

	if pmr.Config != nil {
		blocks := pmr.Config.ToBlocks(stack.OrgName, stack.Id)
		blocksToCreate := map[string]types.Block{}
		for _, cur := range blocks {
			blocksToCreate[cur.Name] = cur
		}
		existing, err := apiClient.Blocks().List(ctx, stack.Id)
		if err != nil {
			fmt.Fprintln(w)
			return fmt.Errorf("error checking for existing blocks: %w", err)
		}
		for _, cur := range existing {
			delete(blocksToCreate, cur.Name)
		}

		if len(blocksToCreate) > 0 {
			colorstring.Fprintf(w, "    [bold]Nullstone will create the following %d blocks...[reset]\n", len(blocksToCreate))
			for name, _ := range blocksToCreate {
				colorstring.Fprintf(w, "        [green]+[reset] %s\n", name)
			}
			if err := resolver.ResourceResolver.BackfillMissingBlocks(ctx, blocks); err != nil {
				fmt.Fprintln(w)
				return fmt.Errorf("error initializing normalization: %w", err)
			}
		} else {
			colorstring.Fprintln(w, "    [green]✔[reset] Nullstone does not need to create any blocks.")
		}
	}

	if errs := iac.Normalize(ctx, pmr, resolver); len(errs) > 0 {
		colorstring.Fprintf(w, "    [bold]Detected errors when validating connections in Nullstone IaC files[reset]\n")
		for _, err := range errs {
			relFilename, _ := filepath.Rel(curDir, err.IacContext.Filename)
			colorstring.Fprintf(w, "        [red]✖[reset] (%s) %s => %s\n", relFilename, err.ObjectPathContext.Context(), err.ErrorMessage)
		}
		fmt.Fprintln(w)
		return fmt.Errorf("IaC files are invalid.")
	} else {
		colorstring.Fprintln(w, "    [green]✔[reset] Connection validation completed successfully.")
	}
	fmt.Fprintln(w)

	return nil
}
