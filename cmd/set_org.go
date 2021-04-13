package cmd

import (
	"errors"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

var SetOrg = cli.Command{
	Name:  "set-org",
	Usage: "Set the organization for the CLI",
	UsageText: `Most Nullstone CLI commands require a configured nullstone organization to operate.
   This command will set the organization for the current profile.
   If you wish to set the organization per command, use the global --org flag instead.`,
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {
		profile, err := config.LoadProfile(GetProfile(c.Parent()))
		if err != nil {
			return err
		}

		if c.NArg() != 1 {
			return errors.New("Usage: nullstone set-org <org-name>")
		}
		return profile.SaveOrg(c.Args().Get(0))
	},
}
