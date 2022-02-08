package cmd

import "github.com/urfave/cli/v2"

var EnvFlag = &cli.StringFlag{
	Name:     "env",
	Usage:    `The environment name.`,
	// TODO: Once we deprecate `nullstone cmd <app> <env>` -> Set this to required
	// Required: true,
}

var EnvOptionalFlag = &cli.StringFlag{
	Name:     "env",
	Usage:    `The environment name.`,
	Required: false,
}
