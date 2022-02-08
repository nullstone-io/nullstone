package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

type ParseBlockEnvFn func(stackName, blockName, envName string) error

func ParseBlockEnv(c *cli.Context, fn ParseAppEnvFn) error {
	blockName := GetApp(c)
	if blockName == "" {
		cli.ShowCommandHelp(c, c.Command.Name)
		return fmt.Errorf("Block Name is required to run this command. Use --block, NULLSTONE_BLOCK env var, or NULLSTONE_APP env var.")
	}

	envName := GetEnvironment(c)
	if envName == "" {
		cli.ShowCommandHelp(c, c.Command.Name)
		return fmt.Errorf("Environment Name is required to run this command. Use --env or NULLSTONE_ENV env var.")
	}
	stackName := GetStack(c)

	return fn(stackName, blockName, envName)
}
