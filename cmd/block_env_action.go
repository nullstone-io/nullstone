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
	"os/signal"
	"syscall"
)

type BlockEnvActionFn func(ctx context.Context, cfg api.Config, stack types.Stack, block types.Block, env types.Environment) error

func BlockEnvAction(c *cli.Context, fn BlockEnvActionFn) error {
	_, cfg, err := SetupProfileCmd(c)
	if err != nil {
		return err
	}

	if c.NArg() != 2 {
		cli.ShowCommandHelp(c, c.Command.Name)
		return fmt.Errorf("invalid usage")
	}
	blockName := c.Args().Get(0)
	envName := c.Args().Get(1)
	stackName := c.String("stack-name")
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

	// Handle Ctrl+C, kill stream
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-term
		cancelFn()
	}()

	return fn(ctx, cfg, sbe.Stack, sbe.Block, sbe.Env)
}
