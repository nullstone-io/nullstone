package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/nullstone.v0/config"
)

var Configure = cli.Command{
	Name: "configure",
	Action: func(c *cli.Context) error {
		fmt.Print("Enter API Key: ")
		bytePassword, err := terminal.ReadPassword(0)
		if err != nil {
			return fmt.Errorf("error reading password: %w", err)
		}
		fmt.Println()
		if err := config.SaveApiKey(string(bytePassword)); err != nil {
			return fmt.Errorf("unable to save API key: %w", err)
		}
		fmt.Println("nullstone configured successfully!")
		return nil
	},
}
