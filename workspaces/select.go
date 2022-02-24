package workspaces

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/git"
	"path"
	"strings"
)

var (
	backendFilename         = "__backend__.tf"
	activeWorkspaceFilename = path.Join(".nullstone", "active-workspace.yml")
)

func Select(ctx context.Context, cfg api.Config, workspace Manifest, runConfig types.RunConfig) error {
	repo := git.RepoFromDir(".")
	if repo != nil {
		// Add gitignores for __backend__.tf and .nullstone/active-workspace.yml
		_, missing := git.FindGitIgnores(repo, []string{
			backendFilename,
			activeWorkspaceFilename,
		})
		if len(missing) > 0 {
			fmt.Printf("Adding %s to .gitignore\n", strings.Join(missing, ", "))
			git.AddGitIgnores(repo, missing)
		}
	}

	if err := WriteBackendTf(cfg, workspace.WorkspaceUid, backendFilename); err != nil {
		return fmt.Errorf("error writing terraform backend file: %w", err)
	}
	if err := workspace.WriteToFile(activeWorkspaceFilename); err != nil {
		return fmt.Errorf("error writing active workspace file: %w", err)
	}

	fmt.Printf(`Selected workspace:
  Stack:     %s
  Block:     %s
  Env:       %s
  Workspace: %s
`, workspace.StackName, workspace.BlockName, workspace.EnvName, workspace.WorkspaceUid)

	capGenerator := CapabilitiesGenerator{
		RegistryAddress:  cfg.BaseAddress,
		Manifest:         workspace,
		TemplateFilename: "capabilities.tf.tmpl",
		TargetFilename:   "capabilities.tf",
	}
	if capGenerator.ShouldGenerate() {
		fmt.Printf("Generating %q from %q\n", capGenerator.TargetFilename, capGenerator.TemplateFilename)
		if err := capGenerator.Generate(runConfig); err != nil {
			return fmt.Errorf("Could not generate %q: %w", capGenerator.TargetFilename, err)
		}
	}

	if err := Init(ctx); err != nil {
		fallbackMessage := `Unable to initialize terraform.
Reset .terraform/ directory and run 'terraform init'.`
		fmt.Println(fallbackMessage)
	}
	return nil
}
