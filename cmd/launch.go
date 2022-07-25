package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"time"
)

// Launch command performs push, deploy, and logs
var Launch = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "launch",
		Usage:     "Launch application (push + deploy + wait-healthy)",
		UsageText: "nullstone launch [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppSourceFlag,
			AppVersionFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, details app.Details) error {
				source, version := c.String("source"), DetectAppVersion(c)
				osWriters := logging.StandardOsWriters{}
				stdout := osWriters.Stdout()
				factory := providers.FindFactory(*details.Module)
				if factory == nil {
					return fmt.Errorf("launch is not supported for this app")
				}

				fmt.Fprintln(stdout, "Pushing app artifact...")
				pusher, err := factory.NewPusher(osWriters, cfg, details)
				if err != nil {
					return fmt.Errorf("error creating app pusher: %w", err)
				} else if pusher == nil {
					return fmt.Errorf("push is not supported for this app")
				}
				if err := pusher.Push(ctx, source, version); err != nil {
					return fmt.Errorf("error pushing artifact: %w", err)
				}
				fmt.Fprintln(stdout, "")

				fmt.Fprintln(stdout, "Deploying app...")
				deployer, err := factory.NewDeployer(osWriters, cfg, details)
				if err != nil {
					return fmt.Errorf("error creating app deployer: %w", err)
				} else if deployer == nil {
					return fmt.Errorf("deploy is not supported for this app")
				}
				reference, err := deployer.Deploy(ctx, version)
				if err != nil {
					return fmt.Errorf("error deploying app: %w", err)
				} else if reference == "" {
					return nil
				}
				fmt.Fprintln(stdout, "")

				fmt.Fprintln(stdout, "Waiting for app to become healthy...")
				deployStatusGetter, err := factory.NewDeployStatusGetter(osWriters, cfg, details)
				if err != nil {
					return fmt.Errorf("error creating app deployment status analyzer: %w", err)
				} else if deployStatusGetter == nil {
					// If we don't have a way of retrieving status, we are just going to complete the deployment
					return nil
				}
				if err := waitHealthy(ctx, deployStatusGetter, reference); err != nil {
					return err
				}
				fmt.Fprintln(stdout, "")
				return nil
			})
		},
	}
}

func waitHealthy(ctx context.Context, deployStatusGetter app.DeployStatusGetter, reference string) error {
	for {
		status, err := deployStatusGetter.GetDeployStatus(ctx, reference)
		if err != nil {
			return fmt.Errorf("error querying app deployment status: %w", err)
		}
		switch status {
		case app.RolloutStatusComplete:
			return nil
		case app.RolloutStatusInProgress:
			return nil
		case app.RolloutStatusFailed:
			return fmt.Errorf("deployment failed")
		case app.RolloutStatusUnknown:
			return fmt.Errorf("unknown app deployment status")
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("cancelled")
		case <-time.After(5 * time.Second):
		}
	}
}
