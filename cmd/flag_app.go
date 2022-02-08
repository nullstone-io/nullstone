package cmd

import "github.com/urfave/cli/v2"

var GlobalAppFlag = &cli.StringFlag{
	Name:    "app",
	Usage:   "Set the application name for commands that require an application.",
	EnvVars: []string{"NULLSTONE_APP"},
}

func GetApp(c *cli.Context) string {
	appName := c.String(GlobalAppFlag.Name)
	// TODO: Drop parsing of first command arg as app once fully deprecated
	if appName == "" && c.NArg() >= 1 {
		appName = c.Args().Get(0)
	}
	return appName
}
