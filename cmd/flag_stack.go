package cmd

import "github.com/urfave/cli/v2"

var GlobalStackFlag = &cli.StringFlag{
	Name: "stack",
	Usage: `Set the stack name that owns the app/block.
       This is only required if multiple apps/blocks have the same name.`,
	EnvVars: []string{"NULLSTONE_STACK"},
}

func GetStack(c *cli.Context) string {
	return c.String(GlobalStackFlag.Name)
}
