package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"os"
	"syscall"
)

var (
	AddressFlag = &cli.StringFlag{
		Name:  "address",
		Value: api.DefaultAddress,
		Usage: "Specify the url for the Nullstone API.",
	}
	ApiKeyFlag = &cli.StringFlag{
		Name:  "api-key",
		Value: "",
		Usage: "Configure your personal API key that will be used to authenticate with the Nullstone API. You can generate an API key from your profile page.",
	}
)

var Configure = &cli.Command{
	Name:        "configure",
	Description: "Establishes a profile and configures authentication for the CLI to use.",
	Usage:       "Configure the nullstone CLI",
	UsageText:   "nullstone configure --api-key=<api-key>",
	Flags: []cli.Flag{
		AddressFlag,
		ApiKeyFlag,
	},
	Action: func(c *cli.Context) error {
		apiKey := c.String(ApiKeyFlag.Name)
		if apiKey == "" {
			fmt.Print("Enter API Key: ")
			rawApiKey, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("error reading password: %w", err)
			}
			fmt.Println()
			apiKey = string(rawApiKey)
		}

		profile := config.Profile{
			Name:    GetProfile(c),
			Address: c.String(AddressFlag.Name),
			ApiKey:  apiKey,
		}
		if err := profile.Save(); err != nil {
			return fmt.Errorf("error configuring profile: %w", err)
		}
		fmt.Fprintln(os.Stderr, "nullstone configured successfully!")
		return nil
	},
}
