package cmd

import "github.com/urfave/cli/v2"

var AppFlag = &cli.StringFlag{
	Name:     "app",
	Usage:    `The application name.`,
	// TODO: Once we deprecate `nullstone cmd <app> <env>` -> Set this to required
	// Required: true,
}
