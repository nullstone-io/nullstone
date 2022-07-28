package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
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
				osWriters := logging.StandardOsWriters{}
				provider := providers.FindFactory(*appDetails.Module)
				if provider == nil {
					return fmt.Errorf("deploy is not supported for this app")
				}
				if err := CreateDeploy(cfg, appDetails, version); err != nil {
					return err
				}
				reference, err := deploy(ctx, cfg, appDetails, osWriters, provider, version)
				if err != nil {
					return err
				}
				if wait {
					if err := waitHealthy(ctx, cfg, appDetails, osWriters, provider, reference); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
}

func deploy(ctx context.Context, cfg api.Config, appDetails app.Details, osWriters logging.OsWriters, provider *app.Provider, version string) (string, error) {
	if provider.NewDeployer == nil {
		return "", fmt.Errorf("This app does not support deploy.")
	}

	stdout := osWriters.Stdout()
	fmt.Fprintln(stdout, "Deploying app...")
	deployer, err := provider.NewDeployer(osWriters, cfg, appDetails)
	if err != nil {
		return "", fmt.Errorf("error creating app deployer: %w", err)
	} else if deployer == nil {
		return "", fmt.Errorf("deploy is not supported for this app")
	}
	reference, err := deployer.Deploy(ctx, version)
	if err != nil {
		return "", fmt.Errorf("error deploying app: %w", err)
	} else if reference == "" {
		return "", nil
	}
	fmt.Fprintln(stdout, "App deployed.")
	fmt.Fprintf(osWriters.Stdout(), "Deploy ID: %s\n", reference)
	fmt.Fprintln(stdout, "")
	return reference, nil
}
