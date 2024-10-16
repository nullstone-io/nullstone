package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	iac2 "gopkg.in/nullstone-io/nullstone.v0/iac"
	"os"
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
				pmr, err := iac2.Discover(curDir, stdout)
				if err != nil {
					return err
				}

				if err := iac2.Process(ctx, cfg, curDir, stdout, *stack, *env, *pmr); err != nil {
					return err
				}

				return iac2.Apply(ctx, cfg, curDir, stdout, *stack, *env, *pmr)
			})
		})
	},
}
