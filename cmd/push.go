package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

// Push command performs a docker push to an authenticated image registry configured against an app/container
var Push = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "push",
		Usage:     "Push artifact",
		UsageText: "nullstone push [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppSourceFlag,
			AppVersionFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				source, version := c.String("source"), DetectAppVersion(c)
				osWriters := logging.StandardOsWriters{}
				provider := providers.FindFactory(*appDetails.Module)
				if provider == nil {
					return fmt.Errorf("push is not supported for this app")
				}
				return push(ctx, cfg, appDetails, osWriters, provider, source, version)
			})
		},
	}
}

func push(ctx context.Context, cfg api.Config, appDetails app.Details, osWriters logging.OsWriters, provider *app.Provider, source, version string) error {
	if provider.NewPusher == nil {
		return fmt.Errorf("This app does not support push.")
	}
	pusher, err := provider.NewPusher(osWriters, cfg, appDetails)
	if err != nil {
		return fmt.Errorf("error creating app pusher: %w", err)
	} else if pusher == nil {
		return fmt.Errorf("this CLI does not support application category=%s, type=%s", appDetails.Module.Category, appDetails.Module.Type)
	}
	stdout := osWriters.Stdout()
	fmt.Fprintln(stdout, "Pushing app artifact...")
	if err := pusher.Push(ctx, source, version); err != nil {
		return fmt.Errorf("error pushing artifact: %w", err)
	}
	fmt.Fprintln(stdout, "App artifact pushed.")
	fmt.Fprintln(stdout, "")
	return nil
}
