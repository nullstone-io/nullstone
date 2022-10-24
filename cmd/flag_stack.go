package cmd

import "github.com/urfave/cli/v2"

var StackFlag = &cli.StringFlag{
	Name:    "stack",
	Usage:   "Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name.",
	EnvVars: []string{"NULLSTONE_STACK"},
}

var StackRequiredFlag = &cli.StringFlag{
	Name:     "stack",
	Usage:    "Name of the stack to use for this operation",
	EnvVars:  []string{"NULLSTONE_STACK"},
	Required: true,
}
