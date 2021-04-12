package cmd

import "github.com/urfave/cli"

var ProfileFlag = cli.StringFlag{
	Name:   "profile",
	EnvVar: "NULLSTONE_PROFILE",
	Value:  "default",
	Usage:  "Name of profile",
}

func GetProfile(c *cli.Context) string {
	val := c.GlobalString(ProfileFlag.Name)
	if val == "" {
		return ProfileFlag.Value
	}
	return val
}
