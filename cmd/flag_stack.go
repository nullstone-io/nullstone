package cmd

import "github.com/urfave/cli/v2"

var StackFlag = &cli.StringFlag{
	Name: "stack",
	Usage: `The stack name where the app resides.
       This is only required if multiple apps have the same 'app-name'.`,
}
