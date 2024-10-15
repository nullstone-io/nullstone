package cmd

import (
	"context"
	"fmt"
	"github.com/mitchellh/colorstring"
	"github.com/nullstone-io/iac"
	"github.com/nullstone-io/iac/core"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/git"
	"os"
	"path/filepath"
)

var Iac = &cli.Command{
	Name:      "iac",
	Usage:     "Utility functions to interact with Nullstone IaC",
	UsageText: "nullstone iac [subcommand]",
	Subcommands: []*cli.Command{
		IacTest,
	},
}

var IacTest = &cli.Command{
	Name:        "test",
	Description: "Test the current repository's IaC files against a Nullstone stack.",
	Usage:       "Test Nullstone IaC",
	UsageText:   "nullstone iac test --stack=<stack> --env=<env>",
	Flags: []cli.Flag{
		StackFlag,
		EnvFlag,
	},
	Action: func(c *cli.Context) error {
		return CancellableAction(func(ctx context.Context) error {
			return ProfileAction(c, func(cfg api.Config) error {
				curDir, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot retrieve Nullstone IaC files: %w", err)
				}
				apiClient := &api.Client{Config: cfg}
				pmr, err := parseIacFiles(curDir)
				if err != nil {
					return err
				}

				stackName := c.String("stack")
				stack, err := find.Stack(ctx, cfg, stackName)
				if err != nil {
					return err
				} else if stack == nil {
					return find.StackDoesNotExistError{StackName: stackName}
				}

				envName := c.String("env")
				env, err := find.Env(ctx, cfg, stack.Id, envName)
				if err != nil {
					return err
				} else if env == nil {
					return find.EnvDoesNotExistError{StackName: stackName, EnvName: envName}
				}
				stdout := os.Stdout

				// Emit information about detected IaC files
				numFiles := len(pmr.Overrides)
				if pmr.Config != nil {
					numFiles++
				}
				colorstring.Fprintf(stdout, "[bold]Found %d IaC files[reset]\n", numFiles)
				if cur := pmr.Config; cur != nil {
					relFilename, _ := filepath.Rel(curDir, cur.IacContext.Filename)
					fmt.Fprintf(stdout, "\t📂 %s\n", relFilename)
				}
				for _, cur := range pmr.Overrides {
					relFilename, _ := filepath.Rel(curDir, cur.IacContext.Filename)
					fmt.Fprintf(stdout, "\t📂 %s\n", relFilename)
				}
				fmt.Fprintln(stdout)

				colorstring.Fprintf(stdout, "[bold]Testing Nullstone IaC files against %s/%s environment...[reset]\n", stack.Name, env.Name)
				resolver := core.NewApiResolver(apiClient, stack.Id, env.Id)
				if errs := iac.Resolve(ctx, *pmr, resolver); len(errs) > 0 {
					colorstring.Fprintf(stdout, "[bold]Detected errors when resolving Nullstone IaC files[reset]\n")
					for _, err := range errs {
						relFilename, _ := filepath.Rel(curDir, err.IacContext.Filename)
						colorstring.Fprintf(stdout, "\t[red]❌[reset] (%s) %s => %s\n", relFilename, err.ObjectPathContext.Context(), err.ErrorMessage)
					}
					fmt.Fprintln(stdout)
					return fmt.Errorf("IaC files are invalid.")
				} else {
					colorstring.Fprintln(stdout, "\t[green]✔️[reset] Resolution completed successfully.")
				}

				if errs := iac.Validate(*pmr); len(errs) > 0 {
					colorstring.Fprintf(stdout, "\t[bold]Detected errors when validating Nullstone IaC files[reset]\n")
					for _, err := range errs {
						relFilename, _ := filepath.Rel(curDir, err.IacContext.Filename)
						colorstring.Fprintf(stdout, "\t\t[red]❌[reset] (%s) %s => %s\n", relFilename, err.ObjectPathContext.Context(), err.ErrorMessage)
					}
					fmt.Fprintln(stdout)
					return fmt.Errorf("IaC files are invalid.")
				} else {
					colorstring.Fprintln(stdout, "\t[green]✔️[reset] Validation completed successfully.")
				}

				if pmr.Config != nil {
					blocks := pmr.Config.ToBlocks(stack.OrgName, stack.Id)
					blocksToCreate := map[string]types.Block{}
					for _, cur := range blocks {
						blocksToCreate[cur.Name] = cur
					}
					existing, err := apiClient.Blocks().List(ctx, stack.Id)
					if err != nil {
						fmt.Fprintln(stdout)
						return fmt.Errorf("error checking for existing blocks: %w", err)
					}
					for _, cur := range existing {
						delete(blocksToCreate, cur.Name)
					}

					if len(blocksToCreate) > 0 {
						colorstring.Fprintf(stdout, "\t[bold]Nullstone will create the following %d blocks...[reset]\n", len(blocksToCreate))
						for name, _ := range blocksToCreate {
							colorstring.Fprintf(stdout, "\t\t[green]+[reset] %s\n", name)
						}
						if err := resolver.ResourceResolver.BackfillMissingBlocks(ctx, blocks); err != nil {
							fmt.Fprintln(stdout)
							return fmt.Errorf("error initializing normalization: %w", err)
						}
					} else {
						colorstring.Fprintln(stdout, "\t[green]✔️[reset]Nullstone does not need to create any blocks.")
					}
				}

				if errs := iac.Normalize(ctx, *pmr, resolver); len(errs) > 0 {
					colorstring.Fprintf(stdout, "\t[bold]Detected errors when validating connections in Nullstone IaC files[reset]\n")
					for _, err := range errs {
						relFilename, _ := filepath.Rel(curDir, err.IacContext.Filename)
						colorstring.Fprintf(stdout, "\t\t[red]❌[reset] (%s) %s => %s\n", relFilename, err.ObjectPathContext.Context(), err.ErrorMessage)
					}
					fmt.Fprintln(stdout)
					return fmt.Errorf("IaC files are invalid.")
				} else {
					colorstring.Fprintln(stdout, "\t[green]✔️[reset] Connection validation completed successfully.")
				}

				return nil
			})
		})
	},
}

func parseIacFiles(dir string) (*iac.ParseMapResult, error) {
	rootDir, err := git.GetRootDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error looking for repository root directory: %w", err)
	} else if rootDir == "" {
		rootDir = dir
	}

	pmr, err := iac.ParseConfigDir(filepath.Join(rootDir, ".nullstone"))
	if err != nil {
		return nil, fmt.Errorf("error parsing nullstone IaC files: %w", err)
	}
	return pmr, nil
}
