package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"os"
)

var Deploy = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "deploy",
		Usage:     "Deploy application",
		UsageText: "nullstone deploy [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppVersionFlag,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
			},
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				version, wait := DetectAppVersion(c), c.IsSet("wait")
				if version == "" {
					return fmt.Errorf("no version specified, version is required to create a deploy")
				}

				deploy, err := CreateDeploy(cfg, appDetails, version)
				if err != nil {
					return err
				}

				if wait {
					// TODO: We should always stream logs, but if --wait is not specified, we would skip "wait-healthy" phase
					return streamDeployLogs(ctx, cfg, *deploy)
				}
				return nil
			})
		},
	}
}

func streamDeployLogs(ctx context.Context, cfg api.Config, deploy types.Deploy) error {
	fmt.Fprintln(os.Stdout, "Waiting for logs...")
	client := api.Client{Config: cfg}
	msgs, err := client.DeployLiveLogs().Watch(ctx, deploy.StackId, deploy.Id)
	if err != nil {
		return err
	}
	for msg := range msgs {
		if msg.Source == "error" {
			return fmt.Errorf(msg.Content)
		}
		fmt.Fprint(os.Stdout, msg.Content)
	}

	updated, err := client.Deploys().Get(deploy.StackId, deploy.AppId, deploy.EnvId, deploy.Id)
	if err != nil {
		return fmt.Errorf("error retrieving deploy status: %w", err)
	}
	switch updated.Status {
	case types.DeployStatusCancelled:
		return fmt.Errorf("Deploy was cancelled.")
	case types.DeployStatusFailed:
		return fmt.Errorf("Deploy failed to complete: %s", updated.StatusMessage)
	}
	return nil
}
