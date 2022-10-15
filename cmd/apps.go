package cmd

import (
	"fmt"
	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var Apps = &cli.Command{
	Name:      "apps",
	Usage:     "View and modify applications",
	UsageText: "nullstone apps [subcommand]",
	Subcommands: []*cli.Command{
		AppsList,
	},
}

var AppsList = &cli.Command{
	Name:      "list",
	Usage:     "List applications",
	UsageText: "nullstone apps list",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "detail",
			Aliases: []string{"d"},
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			allApps, err := client.Apps().List()
			if err != nil {
				return fmt.Errorf("error listing applications: %w", err)
			}

			if c.IsSet("detail") {
				appDetails := make([]string, len(allApps)+1)
				appDetails[0] = "ID|Name|Reference|Category|Type|Module|Stack|Framework"
				for i, app := range allApps {
					var appCategory types.CategoryName
					var appType string
					if appModule, err := find.Module(cfg, app.ModuleSource); err == nil {
						appCategory = appModule.Category
						appType = appModule.Type
					}
					stack, err := client.Stacks().Get(app.StackId)
					if err != nil {
						return fmt.Errorf("error looking for stack %q: %w", app.StackId, err)
					}
					appDetails[i+1] = fmt.Sprintf("%d|%s|%s|%s|%s|%s|%s|%s", app.Id, app.Name, app.Reference, appCategory, appType, app.ModuleSource, stack.Name, app.Framework)
				}
				fmt.Println(columnize.Format(appDetails, columnize.DefaultConfig()))
			} else {
				for _, app := range allApps {
					fmt.Println(app.Name)
				}
			}

			return nil
		})
	},
}
