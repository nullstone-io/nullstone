package cmd

import (
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

// OrgFlag defines a flag that the CLI uses
//   to contextualize API calls by that organization within Nullstone
// The organization takes the following precedence:
//   `--org` flag
//   `NULLSTONE_ORG` env var
//   `~/.nullstone/org` file
var OrgFlag = cli.StringFlag{
	Name:   "org",
	EnvVar: "NULLSTONE_ORG",
	Usage:  "Nullstone organization name used to contextualize API calls",
}

func GetOrg(c *cli.Context, profile config.Profile) string {
	val := c.GlobalString(OrgFlag.Name)
	if val == "" {
		val, _ = profile.LoadOrg()
	}
	return val
}
