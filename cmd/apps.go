package cmd

import (
	"context"
	"fmt"

	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
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
	Name:        "list",
	Description: "Shows a list of the applications that you have access to. Set the `--detail` flag to show more details about each application.",
	Usage:       "List applications",
	UsageText:   "nullstone apps list",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "detail",
			Aliases: []string{"d"},
			Usage:   "Use this flag to show the details for each application",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := context.TODO()
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			allApps, err := client.Apps().GlobalList(ctx)
			if err != nil {
				return fmt.Errorf("error listing applications: %w", err)
			}

			if c.IsSet("detail") {
				appDetails := make([]string, len(allApps)+1)
				appDetails[0] = "ID|Name|Reference|Stack|Framework"
				for i, app := range allApps {
					stack, err := client.Stacks().Get(ctx, app.StackId, false)
					if err != nil {
						return fmt.Errorf("error looking for stack %q: %w", app.StackId, err)
					}
					appDetails[i+1] = fmt.Sprintf("%d|%s|%s|%s|%s", app.Id, app.Name, app.Reference, stack.Name, app.Framework)
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
