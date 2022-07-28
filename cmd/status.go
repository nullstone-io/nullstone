package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"io"
	"time"
)

var (
	defaultWatchInterval = 1 * time.Second
)

var Status = func(providers admin.Providers) *cli.Command {
	return &cli.Command{
		Name:      "status",
		Usage:     "Application Status",
		UsageText: "nullstone status [--stack=<stack-name>] --app=<app-name> [--env=<env-name>] [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvOptionalFlag,
			AppVersionFlag,
			&cli.BoolFlag{
				Name:    "watch",
				Aliases: []string{"w"},
			},
		},
		Action: func(c *cli.Context) error {
			watchInterval := -1 * time.Second
			if c.IsSet("watch") {
				watchInterval = defaultWatchInterval
			}

			return ProfileAction(c, func(cfg api.Config) error {
				return ParseAppEnv(c, false, func(stackName, appName, envName string) error {
					return CancellableAction(func(ctx context.Context) error {
						if envName == "" {
							return appStatus(ctx, cfg, providers, watchInterval, stackName, appName)
						} else {
							return appEnvStatus(ctx, cfg, providers, watchInterval, stackName, appName, envName)
						}
					})
				})
			})
		},
	}
}

func appStatus(ctx context.Context, cfg api.Config, providers admin.Providers, watchInterval time.Duration, stackName, appName string) error {
	application, err := find.App(cfg, appName, stackName)
	if err != nil {
		return err
	}

	client := api.Client{Config: cfg}
	envs, err := client.Environments().List(application.StackId)
	if err != nil {
		return fmt.Errorf("error retrieving environments: %w", err)
	}

	if len(envs) < 1 {
		fmt.Println("No environments exist")
		return nil
	}

	return WatchAction(ctx, watchInterval, func(writer io.Writer) error {
		buffer := &TableBuffer{}
		buffer.AddFields("Env", "Infra", "Version")
		for _, env := range envs {
			awi, err := (NsStatus{Config: cfg}).GetAppWorkspaceInfo(application, env)
			if err != nil {
				return fmt.Errorf("error retrieving app workspace (%s/%s): %w", application.Name, env.Name, err)
			}
			cur := map[string]interface{}{
				"Env":     env.Name,
				"Infra":   awi.Status,
				"Version": awi.Version,
			}

			statuser, _ := providers.FindStatuser(logging.StandardOsWriters{}, cfg, awi.AppDetails)
			if statuser != nil {
				if report, err := statuser.Status(ctx); err != nil {
					return fmt.Errorf("error retrieving app status: %w", err)
				} else {
					buffer.AddFields(report.Fields...)
					for k, v := range report.Data {
						cur[k] = v
					}
				}
			}

			buffer.AddRow(cur)
		}
		fmt.Fprintln(writer, buffer.String())
		return nil
	})
}

func appEnvStatus(ctx context.Context, cfg api.Config, providers admin.Providers, watchInterval time.Duration, stackName, appName, envName string) error {
	_, application, env, err := find.StackAppEnv(cfg, stackName, appName, envName)
	if err != nil {
		return err
	}

	return WatchAction(ctx, watchInterval, func(writer io.Writer) error {
		awi, err := (NsStatus{Config: cfg}).GetAppWorkspaceInfo(application, env)
		if err != nil {
			return fmt.Errorf("error retrieving app workspace (%s/%s): %w", application.Name, env.Name, err)
		}
		appDetails := awi.AppDetails

		fmt.Fprintln(writer, fmt.Sprintf("Env: %s\tInfra: %s\tVersion: %s", appDetails.Env.Name, awi.Status, awi.Version))
		fmt.Fprintln(writer)

		var detailReport admin.StatusDetailReports
		statuser, _ := providers.FindStatuser(logging.StandardOsWriters{}, cfg, awi.AppDetails)
		if statuser != nil {
			var err error
			detailReport, err = statuser.StatusDetail(ctx)
			if err != nil {
				return fmt.Errorf("error retrieving app status detail: %w", err)
			}
		}

		// Emit each report starting with the report name as the header
		for _, report := range detailReport {
			// Emit report header
			fmt.Fprintln(writer, report.Name)

			// Emit report records
			buffer := &TableBuffer{}
			for _, record := range report.Records {
				cur := map[string]interface{}{}
				buffer.AddFields(record.Fields...)
				for k, v := range record.Data {
					cur[k] = v
				}
				buffer.AddRow(cur)
			}
			fmt.Fprintln(writer, buffer.String())
			fmt.Fprintln(writer)
		}

		return nil
	})
}
