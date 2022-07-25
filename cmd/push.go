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
				pusher, err := providers.FindPusher(logging.StandardOsWriters{}, cfg, appDetails)
				if err != nil {
					return fmt.Errorf("error creating app pusher: %w", err)
				} else if pusher == nil {
					return fmt.Errorf("this CLI does not support application category=%s, type=%s", appDetails.Module.Category, appDetails.Module.Type)
				}
				return pusher.Push(ctx, c.String("source"), DetectAppVersion(c))
			})
		},
	}
}
