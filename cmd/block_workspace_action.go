package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"log"
)

type BlockWorkspaceActionFn func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment, workspace types.Workspace) error

func BlockWorkspaceAction(c *cli.Context, fn BlockWorkspaceActionFn) error {
	_, cfg, err := SetupProfileCmd(c)
	if err != nil {
		return err
	}

	stackName := c.String(StackFlag.Name)
	blockName := c.String(BlockFlag.Name)
	envName := c.String(EnvFlag.Name)

	sbe, err := find.StackBlockEnvByName(cfg, stackName, blockName, envName)
	if err != nil {
		return err
	}

	logger := log.New(c.App.ErrWriter, "", 0)
	logger.Printf("Performing workspace command (Org=%s, Block=%s, Stack=%s, Env=%s)", cfg.OrgName, sbe.Block.Name, sbe.Stack.Name, sbe.Env.Name)
	logger.Println()

	client := api.Client{Config: cfg}
	workspace, err := client.Workspaces().Get(sbe.Stack.Id, sbe.Block.Id, sbe.Env.Id)
	if err != nil {
		return fmt.Errorf("error looking for workspace: %w", err)
	} else if workspace == nil {
		return fmt.Errorf("workspace not found")
	}

	return CancellableAction(func(ctx context.Context) error {
		return fn(ctx, cfg, sbe.Stack, sbe.Block, sbe.Env, *workspace)
	})
}
