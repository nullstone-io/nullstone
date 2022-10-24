package cmd

import "github.com/urfave/cli/v2"

var EnvFlag = &cli.StringFlag{
	Name:     "env",
	Usage:    `Name of the environment to use for this operation`,
	EnvVars:  []string{"NULLSTONE_ENV"},
	Required: true,
}

var OldEnvFlag = &cli.StringFlag{
	Name:    "env",
	Usage:   `Name of the environment to use for this operation`,
	EnvVars: []string{"NULLSTONE_ENV"},
	// TODO: Set to required once we fully deprecate parsing app as first command arg
	// Required: true,
}

var EnvOptionalFlag = &cli.StringFlag{
	Name:     "env",
	Usage:    `Name of the environment to use for this operation`,
	EnvVars:  []string{"NULLSTONE_ENV"},
	Required: false,
}

func GetEnvironment(c *cli.Context) string {
	envName := c.String(OldEnvFlag.Name)
	// TODO: Drop parsing of second command arg as env once fully deprecated
	if envName == "" && c.NArg() >= 2 {
		envName = c.Args().Get(1)
	}
	return envName
}
