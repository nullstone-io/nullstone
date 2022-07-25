package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"log"
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
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				osWriters := logging.StandardOsWriters{}
				deployer, err := providers.FindDeployer(osWriters, cfg, appDetails)
				if err != nil {
					return fmt.Errorf("error creating app deployer: %w", err)
				} else if deployer == nil {
					return fmt.Errorf("this CLI does not support application category=%s, type=%s", appDetails.Module.Category, appDetails.Module.Type)
				}
				reference, err := deployer.Deploy(ctx, DetectAppVersion(c))
				if err != nil {
					return err
				}

				fmt.Fprintln(osWriters.Stdout(), reference)
				return nil
			})
		},
	}
}

// TODO: Migrate CLI to use CreateDeploy instead of performing locally?
func CreateDeploy(nsConfig api.Config, stackId, appId, envId int64, version string) error {
	if version == "" {
		return fmt.Errorf("no version specified, version is required to create a deploy")
	}

	client := api.Client{Config: nsConfig}
	result, err := client.Deploys().Create(stackId, appId, envId, version)
	if err != nil {
		return fmt.Errorf("error updating app version: %w", err)
	} else if result == nil {
		return fmt.Errorf("could not find application environment")
	}

	log.Println("Deployment created")
	return nil
}
