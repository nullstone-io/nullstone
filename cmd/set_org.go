package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"os"
)

var SetOrg = &cli.Command{
	Name:        "set-org",
	Description: "Most Nullstone CLI commands require a configured nullstone organization to operate. This command will set the organization for the current profile. If you wish to set the organization per command, use the global `--org` flag instead.",
	Usage:       "Set the organization for the CLI",
	UsageText:   `nullstone set-org <org-name>`,
	Flags:       []cli.Flag{},
	Action: func(c *cli.Context) error {
		profile, err := config.LoadProfile(GetProfile(c))
		if err != nil {
			return err
		}

		if c.NArg() != 1 {
			return cli.ShowCommandHelp(c, "set-org")
		}

		orgName := c.Args().Get(0)
		if err := profile.SaveOrg(orgName); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Organization set to %s for %s profile\n", orgName, profile.Name)
		return nil
	},
}
