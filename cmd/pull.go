package cmd

import (
	"context"
	"fmt"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

// Pull downloads the artifact from an authenticated registry configured against an app/container
var Pull = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:        "push",
		Description: "Download (pull) an artifact containing the source for your application. --version is required to identify which artifact to pull.",
		Usage:       "Pull artifact",
		UsageText:   "nullstone pull [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvFlag,
			AppVersionRequiredFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				osWriters := CliOsWriters{Context: c}
				version := c.String(AppVersionRequiredFlag.Name)

				pusher, err := getPusher(providers, cfg, appDetails)
				if err != nil {
					return err
				}

				fmt.Fprintln(osWriters.Stderr(), "Pulling app artifact...")
				if err := pusher.Pull(ctx, version); err != nil {
					return fmt.Errorf("error pulling artifact: %w", err)
				}
				fmt.Fprintln(osWriters.Stderr(), "App artifact pulled.")
				fmt.Fprintln(osWriters.Stderr(), "")

				return nil
			})
		},
	}
}
