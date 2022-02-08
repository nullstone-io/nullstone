package cmd

import "github.com/urfave/cli/v2"

var GlobalEnvFlag = &cli.StringFlag{
	Name:    "env",
	Usage:   `Set the environment name for commands that require an environment.`,
	EnvVars: []string{"NULLSTONE_ENV"},
}

func GetEnvironment(c *cli.Context) string {
	envName := c.String(GlobalEnvFlag.Name)
	// TODO: Drop parsing of second command arg as env once fully deprecated
	if envName == "" && c.NArg() >= 2 {
		envName = c.Args().Get(1)
	}
	return envName
}

var EnvOptionalFlag = &cli.StringFlag{
	Name:     "env",
	Usage:    `The environment name.`,
	Required: false,
}
