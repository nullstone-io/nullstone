package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nullstone-io/iac"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/git"
	iac2 "gopkg.in/nullstone-io/nullstone.v0/iac"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
)

var Iac = &cli.Command{
	Name:      "iac",
	Usage:     "Utility functions to interact with Nullstone IaC",
	UsageText: "nullstone iac [subcommand]",
	Subcommands: []*cli.Command{
		IacTest,
		IacGenerate,
		IacSync,
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
				pmr, err := iac2.Discover(curDir, stackName, stdout)
				if err != nil {
					return err
				}

				if err := iac2.Process(ctx, cfg, curDir, stdout, *stack, *env, *pmr); err != nil {
					return err
				}

				return iac2.Test(ctx, cfg, stdout, *stack, *env, *pmr)
			})
		})
	},
}

var IacSync = &cli.Command{
	Name:        "sync",
	Description: "Sync IaC configuration to a Nullstone environment and optionally trigger infra updates.",
	Usage:       "Sync Nullstone IaC",
	UsageText:   "nullstone iac sync --stack=<stack> --env=<env> [--auto-plan] [--auto-apply] [--from-git]",
	Flags: []cli.Flag{
		StackFlag,
		EnvFlag,
		&cli.BoolFlag{
			Name:  "auto-plan",
			Usage: "Queue an infra-update Run on each workspace where IaC changes are detected. The Run is left pending approval.",
		},
		&cli.BoolFlag{
			Name:  "auto-apply",
			Usage: "Auto-approve any infra-update Run created by the sync. Implies --auto-plan.",
		},
		&cli.BoolFlag{
			Name:  "from-git",
			Usage: "Force the server to fetch IaC files from the connected GitHub repo at the resolved commit SHA. By default the local files (already discovered by this command) are submitted in the request payload.",
		},
	},
	Action: func(c *cli.Context) error {
		return CancellableAction(func(ctx context.Context) error {
			return ProfileAction(c, func(cfg api.Config) error {
				stackName := c.String(StackFlag.Name)
				envName := c.String(EnvFlag.Name)

				stack, err := find.Stack(ctx, cfg, stackName)
				if err != nil {
					return err
				} else if stack == nil {
					return find.StackDoesNotExistError{StackName: stackName}
				}
				env, err := find.Env(ctx, cfg, stack.Id, envName)
				if err != nil {
					return err
				} else if env == nil {
					return find.EnvDoesNotExistError{StackName: stackName, EnvName: envName}
				}

				curDir, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot retrieve Nullstone IaC files: %w", err)
				}
				stdout := os.Stdout

				// Run the same Discover → Process → Test pipeline as `nullstone iac test`
				// so the user gets immediate feedback if the IaC files are invalid.
				// Bail out before hitting the API if validation fails.
				pmr, err := iac2.Discover(curDir, stackName, stdout)
				if err != nil {
					return err
				}
				if err := iac2.Process(ctx, cfg, curDir, stdout, *stack, *env, *pmr); err != nil {
					return err
				}
				if err := iac2.Test(ctx, cfg, stdout, *stack, *env, *pmr); err != nil {
					return err
				}

				// Best-effort commit info from the local git repo. Non-git CWDs (CI workspaces,
				// ephemeral runners) get an empty struct — that is fine; the server treats
				// CommitInfo as advisory.
				commitInfo, err := vcs.GetCommitInfo()
				if err != nil {
					fmt.Fprintf(stdout, "warning: could not read git repo (%v); continuing without commit info\n", err)
					commitInfo = types.CommitInfo{}
				}

				autoApply := c.Bool("auto-apply")
				payload := api.TriggerIacSyncPayload{
					AutoPlan:   c.Bool("auto-plan") || autoApply,
					AutoApply:  autoApply,
					CommitInfo: commitInfo,
				}

				// Detached mode: submit the locally-discovered IaC files in the payload
				// so the server doesn't re-fetch from GitHub. --from-git opts back into
				// the legacy fetch path (useful for syncing a remote SHA the user doesn't
				// have locally).
				if !c.Bool("from-git") {
					files, err := readDiscoveredIacFiles(curDir, pmr)
					if err != nil {
						return fmt.Errorf("error reading discovered IaC files: %w", err)
					}
					payload.YamlConfigFiles = files
				}

				client := api.Client{Config: cfg}
				wf, err := client.IacSyncs().Trigger(ctx, stack.Id, env.Id, payload)
				if err != nil {
					return err
				}
				shaSuffix := ""
				if len(commitInfo.CommitSha) >= 8 {
					shaSuffix = " @ " + commitInfo.CommitSha[:8]
				}
				mode := "detached"
				if len(payload.YamlConfigFiles) == 0 {
					mode = "git-fetch"
				}
				fmt.Fprintf(stdout, "Triggered IaC sync (intent workflow %d, %s) for %s/%s%s\n",
					wf.Id, mode, stackName, envName, shaSuffix)
				return nil
			})
		})
	},
}

// readDiscoveredIacFiles reads each file referenced by `iac.Discover`'s ConfigFiles result
// (the Config plus every Overrides entry) and returns them keyed by repo-relative path.
// We deliberately reuse the discovered set rather than walking the working tree — Discover
// is the canonical authority on which files belong in the sync.
//
// rootDir resolution:
//   - When the working directory is inside a git repo, keys are relative to the repo root,
//     producing stable paths like ".nullstone/dev.yml" that match what the GitHub-fetch path
//     would produce.
//   - When there is no git repo (CI ephemeral checkouts, etc.), keys fall back to being
//     relative to curDir.
func readDiscoveredIacFiles(curDir string, pmr *iac.ConfigFiles) (map[string]string, error) {
	if pmr == nil {
		return nil, nil
	}
	rootDir, _, err := git.GetRootDir(curDir)
	if err != nil {
		return nil, fmt.Errorf("error resolving git root: %w", err)
	}
	if rootDir == "" {
		rootDir = curDir
	}

	files := map[string]string{}
	add := func(filename string) error {
		if filename == "" {
			return nil
		}
		key, err := filepath.Rel(rootDir, filename)
		if err != nil {
			return fmt.Errorf("cannot compute repo-relative path for %s: %w", filename, err)
		}
		// Normalize Windows separators so map keys are stable across platforms — the server
		// treats keys as opaque labels but downstream code compares against `.nullstone/...`
		// strings produced by the GitHub-fetch path on Linux.
		key = strings.ReplaceAll(key, string(filepath.Separator), "/")
		raw, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", filename, err)
		}
		files[key] = string(raw)
		return nil
	}

	if pmr.Config != nil {
		if err := add(pmr.Config.IacContext.Filename); err != nil {
			return nil, err
		}
	}
	for _, override := range pmr.Overrides {
		if override == nil {
			continue
		}
		if err := add(override.IacContext.Filename); err != nil {
			return nil, err
		}
	}
	return files, nil
}

var IacGenerate = &cli.Command{
	Name:        "generate",
	Description: "Generate IaC from a Nullstone stack for apps",
	Usage:       "Generate IaC from an application workspace",
	UsageText:   "nullstone iac --stack=<stack> --env=<env> --app=<app>",
	Flags: []cli.Flag{
		StackFlag,
		EnvFlag,
		&cli.StringFlag{
			Name:    "block",
			Usage:   "Name of the block to use for this operation",
			EnvVars: []string{"NULLSTONE_BLOCK"},
		},
	},
	Action: func(c *cli.Context) error {
		return BlockWorkspaceAction(c, func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, workspace types.Workspace) error {
			apiClient := api.Client{Config: cfg}
			buf := bytes.NewBufferString("")
			err := apiClient.WorkspaceConfigFiles().GetConfigFile(ctx, stack.Id, block.Id, env.Id, buf)
			if err != nil {
				return err
			}

			fmt.Fprintln(os.Stdout, buf.String())
			return nil
		})
	},
}
