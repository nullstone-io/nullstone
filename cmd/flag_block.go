package cmd

import "github.com/urfave/cli/v2"

var BlockFlag = &cli.StringFlag{
	Name:     "block",
	Usage:    "Name of the block to use for this operation",
	EnvVars:  []string{"NULLSTONE_BLOCK", "NULLSTONE_APP"},
	Required: true,
}
