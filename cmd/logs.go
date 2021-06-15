package cmd

import (
	"context"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/app_logs"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"os"
	"time"
)

var Logs = func(providers app.Providers, logProviders app_logs.Providers) *cli.Command {
	return &cli.Command{
		Name:      "logs",
		Usage:     "Emit application logs",
		UsageText: "nullstone logs [options] <app-name> <env-name>",
		Flags: []cli.Flag{
			StackFlag,
			&cli.DurationFlag{
				Name:        "start-time",
				Aliases:     []string{"s"},
				DefaultText: "1h",
				Usage: `
       Emit log events that occur after the specified start-time. 
       This is a golang duration relative to the time the command is issued.
       Examples: '5s' (5 seconds ago), '1m' (1 minute ago), '24h' (24 hours ago)
      `,
			},
			&cli.DurationFlag{
				Name:    "end-time",
				Aliases: []string{"e"},
				Usage: `
       Emit log events that occur before the specified end-time. 
       This is a golang duration relative to the time the command is issued.
       Examples: '5s' (5 seconds ago), '1m' (1 minute ago), '24h' (24 hours ago)
      `,
			},
			&cli.DurationFlag{
				Name:        "interval",
				DefaultText: "1s",
				Usage: `Set --interval to a golang duration to control how often to pull new log events.
       This will do nothing unless --tail is set.
      `,
			},
			&cli.BoolFlag{
				Name:    "tail",
				Aliases: []string{"t"},
				Usage: `Set tail to watch log events and emit as they are reported.
       Use --interval to control how often to query log events.
       This is off by default, command will exit as soon as current log events are emitted.`,
			},
		},
		Action: func(c *cli.Context) error {
			return AppAction(c, providers, func(ctx context.Context, cfg api.Config, provider app.Provider, details app.Details) error {
				logStreamOptions := config.LogStreamOptions{
					WatchInterval: -1 * time.Second, // Disabled by default
					Out:           os.Stdout,
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
					logStreamOptions.WatchInterval = time.Duration(0)
					if c.IsSet("interval") {
						logStreamOptions.WatchInterval = c.Duration("interval")
					}
				}

				logProvider, err := logProviders.Identify(provider.DefaultLogProvider(), cfg, details)
				if err != nil {
					return err
				}
				return logProvider.Stream(ctx, cfg, details, logStreamOptions)
			})
		},
	}
}
