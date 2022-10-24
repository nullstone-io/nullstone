package cmd

import "github.com/urfave/cli/v2"

var ProfileFlag = &cli.StringFlag{
	Name:    "profile",
	EnvVars: []string{"NULLSTONE_PROFILE"},
	Value:   "default",
	Usage:   "Name of the profile to use for the operation",
}

func GetProfile(c *cli.Context) string {
	val := c.String(ProfileFlag.Name)
	if val == "" {
		return ProfileFlag.Value
	}
	return val
}
