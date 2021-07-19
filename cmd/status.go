package cmd

import (
	"context"
	"fmt"
	"github.com/gosuri/uilive"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	defaultWatchInterval = 1 * time.Second
)

var Status = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "status",
		Usage:     "Application Status",
		UsageText: "nullstone status [options] <app-name> [<env-name>]",
		Flags: []cli.Flag{
			StackFlag,
			AppVersionFlag,
			&cli.BoolFlag{
				Name:    "watch",
				Aliases: []string{"w"},
			},
		},
		Action: func(c *cli.Context) error {
			return ProfileAction(c, func(cfg api.Config) error {
				var appName, envName string
				stackName := c.String("stack-name")

				ctx := context.Background()
				// Handle Ctrl+C, kill stream
				ctx, cancelFn := context.WithCancel(ctx)
				defer cancelFn()
				term := make(chan os.Signal, 1)
				signal.Notify(term, os.Interrupt, syscall.SIGTERM)
				go func() {
					<-term
					cancelFn()
				}()

				watchInterval := -1 * time.Second
				if c.IsSet("watch") {
					watchInterval = defaultWatchInterval
				}

				switch c.NArg() {
				case 1:
					appName = c.Args().Get(0)
					return appStatus(ctx, cfg, providers, watchInterval, stackName, appName)
				case 2:
					appName = c.Args().Get(0)
					envName = c.Args().Get(1)
					return appEnvStatus(ctx, cfg, providers, watchInterval, stackName, appName, envName)
				default:
					cli.ShowCommandHelp(c, c.Command.Name)
					return fmt.Errorf("invalid usage")
				}
			})
		},
	}
}

func appStatus(ctx context.Context, cfg api.Config, providers app.Providers, watchInterval time.Duration, stackName, appName string) error {
	finder := NsFinder{Config: cfg}
	application, _, err := finder.FindAppAndStack(appName, stackName)
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

	getStatusReport := func(appDetails app.Details) (app.StatusReport, error) {
		var report app.StatusReport

		if appDetails.Workspace.Status == types.WorkspaceStatusNotProvisioned {
			return report, nil
		}

		provider := providers.Find(appDetails.Workspace.Module.Category, appDetails.Workspace.Module.Type)
		if provider == nil {
			return report, nil
		}
		return provider.Status(cfg, appDetails)
	}

	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	for {
		buffer := &TableBuffer{}
		buffer.AddFields("env", "infra", "version")
		for _, env := range envs {
			awi, err := (NsStatus{Config: cfg}).GetAppWorkspaceInfo(application, env)
			if err != nil {
				return fmt.Errorf("error retrieving app workspace (%s/%s): %w", application.Name, env.Name, err)
			}
			cur := map[string]interface{}{
				"env":     env.Name,
				"infra":   awi.Status,
				"version": awi.Version,
			}

			report, err := getStatusReport(awi.AppDetails)
			if err != nil {
				return fmt.Errorf("error retrieving app status: %w", err)
			} else {
				buffer.AddFields(report.Fields...)
				for k, v := range report.Data {
					cur[k] = v
				}
			}
			
			buffer.AddRow(cur)
		}
		fmt.Fprintln(writer, buffer.Serialize("|"))

		if watchInterval <= 0*time.Second {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(watchInterval):
		}
	}
}

func appEnvStatus(ctx context.Context, cfg api.Config, providers app.Providers, watchInterval time.Duration, stackName, appName, envName string) error {
	finder := NsFinder{Config: cfg}
	appDetails, err := finder.FindAppDetails(appName, stackName, envName)
	if err != nil {
		return err
	}

	return nil
}

func getStatusReport(providers app.Providers, details app.Details) (app.StatusReport, error) {
	workspace := details.Workspace
	if workspace.Status == types.WorkspaceStatusNotProvisioned {
		return app.StatusReport{}, nil
	}

	provider := providers.Find(workspace.Module.Category, workspace.Module.Type)
	if provider == nil {
		return report, fmt.Errorf("this CLI does not support application category=%s, type=%s", workspace.Module.Category, workspace.Module.Type)
	}
	return provider.Status(cfg, details)
}
