package cmd

import (
	"errors"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

var (
	ErrMissingOrg = errors.New("An organization has not been configured with this profile. See 'nullstone set-org -h' for more details.")
)

// OrgFlag defines a flag that the CLI uses
//   to contextualize API calls by that organization within Nullstone
// The organization takes the following precedence:
//   `--org` flag
//   `NULLSTONE_ORG` env var
//   `~/.nullstone/<profile>/org` file
var OrgFlag = &cli.StringFlag{
	Name:    "org",
	EnvVars: []string{"NULLSTONE_ORG"},
	Usage:   `Nullstone organization name used to contextualize API calls. If this flag is not specified, the nullstone CLI will use ~/.nullstone/<profile>/org file.`,
}

func GetOrg(c *cli.Context, profile config.Profile) string {
	val := c.String(OrgFlag.Name)
	if val == "" {
		val, _ = profile.LoadOrg()
	}
	return val
}
