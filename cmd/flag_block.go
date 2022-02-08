package cmd

import "github.com/urfave/cli/v2"

var BlockFlag = &cli.StringFlag{
	Name:     "block",
	Usage:    `The block name.`,
	Required: true,
}
