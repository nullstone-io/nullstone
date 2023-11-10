package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

var Profile = &cli.Command{
	Name:      "profile",
	Usage:     "View the current profile and its configuration",
	UsageText: "nullstone profile",
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			fmt.Printf("Profile: %s\n", GetProfile(c))
			fmt.Printf("API Address: %s\n", cfg.BaseAddress)
			accessToken, err := cfg.AccessTokenSource.GetAccessToken(cfg.OrgName)
			if err != nil {
				return err
			}
			if accessToken != "" {
				fmt.Printf("API Key: *** redacted ***\n")
			} else {
				fmt.Printf("API Key: (not set)\n")
			}
			fmt.Printf("Is Trace Enabled: %t\n", cfg.IsTraceEnabled)
			fmt.Printf("Org Name: %s\n", cfg.OrgName)
			fmt.Println()
			return nil
		})
	},
}
