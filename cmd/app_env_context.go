package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

type ParseAppEnvFn func(stackName, appName, envName string) error

func ParseAppEnv(c *cli.Context, isEnvRequired bool, fn ParseAppEnvFn) error {
	stackName := c.String(StackFlag.Name)
	appName := GetApp(c)
	// TODO: Drop this validation once AppFlag.Required=true
	if appName == "" {
		cli.ShowCommandHelp(c, c.Command.Name)
		return fmt.Errorf("App Name is required to run this command. Use --app or NULLSTONE_APP env var.")
	}

	envName := GetEnvironment(c)
	// TODO: Drop this validation once EnvFlag.Required=true
	if isEnvRequired && envName == "" {
		cli.ShowCommandHelp(c, c.Command.Name)
		return fmt.Errorf("Environment Name is required to run this command. Use --env or NULLSTONE_ENV env var.")
	}

	return fn(stackName, appName, envName)
}
