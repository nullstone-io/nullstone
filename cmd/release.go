package cmd

import (
	"context"
	"fmt"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app_urls"
)

var Release = func(providers app.Providers) *cli.Command {
	waitFlag := &cli.BoolFlag{
		Name:    "wait",
		Aliases: []string{"w"},
		Usage:   "Wait for the release to complete and stream the logs to the console.",
	}
	autoApproveFlag := &cli.BoolFlag{
		Name:  "auto-approve",
		Usage: "Skip any approvals on the infra-update. This requires proper permissions in the stack.",
	}

	return &cli.Command{
		Name: "release",
		Description: "Make your infra and app code live through a single workflow. Nullstone runs an infra-update " +
			"when there are outstanding workspace changes and a deploy for the resolved app version, picking the " +
			"optimal, most reliable path.",
		Usage:     "Release infra changes and/or a new app version",
		UsageText: "nullstone release [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppVersionFlag,
			EnvVarFlag,
			autoApproveFlag,
			waitFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				osWriters := CliOsWriters{Context: c}
				wait := c.IsSet(waitFlag.Name)

				info, err := calcDeployInfo(ctx, c)
				if err != nil {
					return err
				}

				envVars, err := ParseEnvVars(c)
				if err != nil {
					return err
				}

				var autoApprove *bool
				if c.IsSet(autoApproveFlag.Name) {
					val := c.Bool(autoApproveFlag.Name)
					autoApprove = &val
				}

				payload := api.ReleaseCreatePayload{
					AutomationTool: detectAutomationTool(),
					IsApproved:     autoApprove,
					Apps: []api.ReleaseApp{
						{
							AppId:      appDetails.App.Id,
							FromSource: false,
							Version:    info.EffectiveVersion,
							CommitSha:  info.CommitInfo.CommitSha,
							EnvVars:    envVars,
						},
					},
				}

				fmt.Fprintln(osWriters.Stderr(), "Creating release...")
				client := api.Client{Config: cfg}
				iw, err := client.Releases().Create(ctx, appDetails.App.StackId, appDetails.Env.Id, payload)
				if err != nil {
					return fmt.Errorf("error creating release: %w", err)
				} else if iw == nil {
					return fmt.Errorf("unable to create release")
				}
				fmt.Fprintln(osWriters.Stderr())

				// Without --wait, print the workflow URL and return; with --wait, stream the logs.
				if !wait {
					fmt.Fprintln(osWriters.Stderr(), "Started release, view status here:")
					fmt.Fprintln(osWriters.Stdout(), app_urls.GetIntentWorkflow(cfg, *iw))
					return nil
				}
				return streamDeployIntentLogs(ctx, osWriters, cfg, appDetails, *iw, wait)
			})
		},
	}
}
