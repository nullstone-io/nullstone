package cmd

import "github.com/urfave/cli/v2"

var GlobalBlockFlag = &cli.StringFlag{
	Name:    "block",
	Usage:   "Set the block name for commands that require a block.",
	EnvVars: []string{"NULLSTONE_BLOCK", "NULLSTONE_APP"},
}

func GetBlock(c *cli.Context) string {
	return c.String(GlobalBlockFlag.Name)
}
