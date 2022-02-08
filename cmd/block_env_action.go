package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"log"
	"os"
)

type BlockEnvActionFn func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment) error

func BlockEnvAction(c *cli.Context, fn BlockEnvActionFn) error {
	_, cfg, err := SetupProfileCmd(c)
	if err != nil {
		return err
	}

	blockName := c.String(BlockFlag.Name)
	envName := c.String(EnvFlag.Name)

	// TODO: Remove this block once we deprecate `nullstone cmd <app> <env>`
	if envName == "" {
		cli.ShowCommandHelp(c, c.Command.Name)
		return fmt.Errorf("'--env' flag is required to run this command")
	}

	stackName := c.String(StackFlag.Name)
	specifiedStack := stackName
	if specifiedStack == "" {
		specifiedStack = "<unspecified>"
	}

	logger := log.New(os.Stderr, "", 0)
	logger.Printf("Performing workspace command (Org=%s, Block=%s, Stack=%s, Env=%s)", cfg.OrgName, blockName, specifiedStack, envName)
	logger.Println()

	sbe, err := find.StackBlockEnvByName(cfg, stackName, blockName, envName)
	if err != nil {
		return err
	}

	return CancellableAction(func(ctx context.Context) error {
		return fn(ctx, cfg, sbe.Stack, sbe.Block, sbe.Env)
	})
}
