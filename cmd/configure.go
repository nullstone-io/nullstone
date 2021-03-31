package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"syscall"
)

var (
	AddressFlag = cli.StringFlag{
		Name:  "address",
		Value: "https://api.nullstone.io",
		Usage: "Nullstone API Address",
	}
)

var Configure = cli.Command{
	Name: "configure",
	Flags: []cli.Flag{
		AddressFlag,
	},
	Action: func(c *cli.Context) error {
		fmt.Print("Enter API Key: ")
		rawApiKey, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("error reading password: %w", err)
		}
		fmt.Println()

		profile := config.Profile{
			Name:    GetProfile(c),
			Address: c.String("address"),
			ApiKey:  string(rawApiKey),
		}
		if err := profile.Save(); err != nil {
			return fmt.Errorf("error configuring profile: %w", err)
		}
		fmt.Println("nullstone configured successfully!")
		return nil
	},
}
