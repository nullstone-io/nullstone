package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
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
	UsageText:   "nullstone iac sync --stack=<stack> --env=<env> [--auto-plan] [--auto-apply]",
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

				client := api.Client{Config: cfg}
				wf, err := client.IacSyncs().Trigger(ctx, stack.Id, env.Id, payload)
				if err != nil {
					return err
				}
				shaSuffix := ""
				if len(commitInfo.CommitSha) >= 8 {
					shaSuffix = " @ " + commitInfo.CommitSha[:8]
				}
				fmt.Fprintf(stdout, "Triggered IaC sync (intent workflow %d) for %s/%s%s\n",
					wf.Id, stackName, envName, shaSuffix)

				// TODO(detached-mode, NUL-49): once the API accepts YamlConfigFiles, reuse
				// the *iac.ConfigFiles already returned by iac2.Discover above to read each
				// file's contents and submit them alongside CommitInfo. The discovered
				// ConfigFiles (Config + Overrides, each carrying IacContext.Filename) is the
				// canonical set of files — do not walk the working tree independently.
				// Design doc: iac-sync-detached-mode.md.
				return nil
			})
		})
	},
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
