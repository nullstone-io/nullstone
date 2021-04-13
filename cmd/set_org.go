package cmd

import (
	"errors"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

var SetOrg = cli.Command{
	Name:  "set-org",
	Usage: "Set the organization for the CLI",
	UsageText: `This commands sets the organization context for the CLI. 
This will create a file at ~/.nullstone/org that is used by the CLI for commands that require an organization.`,
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
