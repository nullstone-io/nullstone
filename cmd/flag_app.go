package cmd

import "github.com/urfave/cli/v2"

var AppFlag = &cli.StringFlag{
	Name:    "app",
	Usage:   "Set the application name.",
	EnvVars: []string{"NULLSTONE_APP"},
	// TODO: Set to required once we fully deprecate parsing app as first command arg
	// Required: true,
}

func GetApp(c *cli.Context) string {
	appName := c.String(AppFlag.Name)
	// TODO: Drop parsing of first command arg as app once fully deprecated
	if appName == "" && c.NArg() >= 1 {
		appName = c.Args().Get(0)
	}
	return appName
}
