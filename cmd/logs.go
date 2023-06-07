package cmd

import (
	"context"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"time"
)

var Logs = func(providers admin.Providers) *cli.Command {
	return &cli.Command{
		Name:        "logs",
		Description: "Streams an application's logs to the console for the given environment. Use the start-time `-s` and end-time `-e` flags to only show logs for a given time period. Use the tail flag `-t` to stream the logs in real time.",
		Usage:       "Emit application logs",
		UsageText:   "nullstone logs [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			&cli.DurationFlag{
				Name:        "start-time",
				Aliases:     []string{"s"},
				DefaultText: "0s",
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
       This will do nothing unless --tail is set. The default is '1s' (1 second).
      `,
			},
			&cli.BoolFlag{
				Name:    "tail",
				Aliases: []string{"t"},
				Usage: `Set tail to watch log events and emit as they are reported.
       Use --interval to control how often to query log events.
       This is off by default. Unless this option is provided, this command will exit as soon as current log events are emitted.`,
			},
		},
		Action: func(c *cli.Context) error {
			logStreamOptions := config.LogStreamOptions{
				WatchInterval: -1 * time.Second, // Disabled by default
			}
			if c.IsSet("start-time") {
				absoluteTime := time.Now().Add(-c.Duration("start-time"))
				logStreamOptions.StartTime = &absoluteTime
			} else {
				absoluteTime := time.Now()
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

			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				logStreamer, err := providers.FindLogStreamer(logging.StandardOsWriters{}, cfg, appDetails)
				if err != nil {
					return err
				}
				return logStreamer.Stream(ctx, logStreamOptions)
			})
		},
	}
}
