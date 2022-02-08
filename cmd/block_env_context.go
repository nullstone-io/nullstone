package cmd

import (
	"github.com/urfave/cli/v2"
)

type ParseBlockEnvFn func(stackName, blockName, envName string) error

func ParseBlockEnv(c *cli.Context, fn ParseAppEnvFn) error {
	stackName := c.String(StackFlag.Name)
	blockName := c.String(BlockFlag.Name)
	envName := c.String(EnvFlag.Name)

	return fn(stackName, blockName, envName)
}
