package cmd

import "github.com/urfave/cli/v2"

var BlockFlag = &cli.StringFlag{
	Name:     "block",
	Usage:    "Set the block name.",
	EnvVars:  []string{"NULLSTONE_BLOCK", "NULLSTONE_APP"},
	Required: true,
}
