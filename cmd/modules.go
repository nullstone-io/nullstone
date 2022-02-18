package cmd

import (
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

var Modules = &cli.Command{
	Name:      "modules",
	Usage:     "View and modify modules",
	UsageText: "nullstone modules [subcommand]",
	Subcommands: []*cli.Command{
		ModulesNew,
	},
}

var ModulesNew = &cli.Command{
	Name:      "new",
	Usage:     "Create new module",
	UsageText: "nullstone modules new",
	Flags:     []cli.Flag{},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			survey := &moduleSurvey{}
			module, err := survey.Ask(cfg)
			if err != nil {
				return err
			}

			client := api.Client{Config: cfg}
			return client.Org(module.OrgName).Modules().Create(module)
		})
	},
}
