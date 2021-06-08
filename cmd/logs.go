package cmd

import (
	"context"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/app_logs"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"time"
)

var Logs = func(providers app.Providers, logProviders app_logs.Providers) cli.Command {
	return cli.Command{
		Name:      "logs",
		Usage:     "Emit application logs",
		UsageText: "nullstone logs <app-name> <env-name> [options]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: "stack",
				Usage: `The stack name where the app resides.
       This is only required if multiple apps have the same 'app-name'.`,
			},
			cli.DurationFlag{
				Name: "start-time",
				Usage: `Emit log events that occur after the specified start-time. 
       This is a golang duration relative to the time the command is issued.
	   Examples: '5s' (5 seconds ago), '1m' (1 minute ago), '24h' (24 hours ago)`,
			},
			cli.DurationFlag{
				Name: "end-time",
				Usage: `Emit log events that occur before the specified end-time. 
       This is a golang duration relative to the time the command is issued.
	   Examples: '5s' (5 seconds ago), '1m' (1 minute ago), '24h' (24 hours ago)`,
			},
			cli.DurationFlag{
				Name: "tail",
				Usage: `Set tail to a golang duration to watch log events and emit as they are reported.
The duration determines how often to query for new log events.
If you specify '--tail', will use the default watch interval of 1 second.
This is off by default, command will exit as soon as current log events are emitted.`,
			},
		},
		Action: func(c *cli.Context) error {
			return AppAction(c, providers, func(ctx context.Context, cfg api.Config, provider app.Provider, details AppDetails) error {
				logStreamOptions := config.LogStreamOptions{
					WatchInterval: -1 * time.Second, // Disabled by default
				}
				if c.IsSet("start-time") {
					absoluteTime := time.Now().Add(-c.Duration("start-time"))
					logStreamOptions.StartTime = &absoluteTime
				}
				if c.IsSet("end-time") {
					absoluteTime := time.Now().Add(-c.Duration("end-time"))
					logStreamOptions.EndTime = &absoluteTime
				}
				if c.IsSet("tail") {
					logStreamOptions.WatchInterval = c.Duration("tail")
				}

				logProvider, err := logProviders.Identify(provider.DefaultLogProvider(), cfg, details.App, details.Workspace)
				if err != nil {
					return err
				}
				return logProvider.Stream(ctx, cfg, details.App, details.Workspace, logStreamOptions)
			})
		},
	}
}
