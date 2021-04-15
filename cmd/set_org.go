package cmd

import (
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"os"
)

var SetOrg = cli.Command{
	Name:  "set-org",
	Usage: "Set the organization for the CLI",
	UsageText: `Most Nullstone CLI commands require a configured nullstone organization to operate.
   This command will set the organization for the current profile.
   If you wish to set the organization per command, use the global --org flag instead.`,
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {
		profile, err := config.LoadProfile(GetProfile(c))
		if err != nil {
			return err
		}

		if c.NArg() != 1 {
			return errors.New("Usage: nullstone set-org <org-name>")
		}

		orgName := c.Args().Get(0)
		if err := profile.SaveOrg(orgName); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Organization set to %s for %s profile\n", orgName, profile.Name)
		return nil
	},
}
