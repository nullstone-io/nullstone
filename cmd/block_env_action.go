package cmd

import (
	"context"
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

	return ParseBlockEnv(c, func(stackName, blockName, envName string) error {
		logger := log.New(os.Stderr, "", 0)
		logger.Printf("Performing workspace command (Org=%s, Block=%s, Stack=%s, Env=%s)", cfg.OrgName, blockName, stackName, envName)
		logger.Println()

		sbe, err := find.StackBlockEnvByName(cfg, stackName, blockName, envName)
		if err != nil {
			return err
		}

		return CancellableAction(func(ctx context.Context) error {
			return fn(ctx, cfg, sbe.Stack, sbe.Block, sbe.Env)
		})
	})
}
