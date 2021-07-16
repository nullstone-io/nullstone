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

func calcPrettyInfraStatus(workspace *types.Workspace) string {
	if workspace == nil {
		return types.WorkspaceStatusNotProvisioned
	}
	if workspace.ActiveRun == nil {
		return workspace.Status
	}
	switch workspace.ActiveRun.Status {
	default:
		return workspace.Status
	case types.RunStatusResolving:
	case types.RunStatusInitializing:
	case types.RunStatusAwaiting:
	case types.RunStatusRunning:
	}
	if workspace.ActiveRun.IsDestroy {
		return "destroying"
	}
	if workspace.Status == types.WorkspaceStatusNotProvisioned {
		return "creating"
	}
	return "updating"
}

func appStatus(ctx context.Context, cfg api.Config, providers app.Providers, watchInterval time.Duration, stackName, appName string) error {
	finder := NsFinder{Config: cfg}
	application, _, err := finder.GetAppAndStack(appName, stackName)
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

	getWorkspaceDetails := func(env *types.Environment) (string, string, app.StatusReport, error) {
		workspace, err := client.Workspaces().Get(application.StackId, application.Id, env.Id)
		if err != nil {
			return "", "", app.StatusReport{}, err
		} else if workspace == nil {
			return types.WorkspaceStatusNotProvisioned, "not-deployed", app.StatusReport{}, nil
		}
		infraStatus := calcPrettyInfraStatus(workspace)

		appEnv, err := client.AppEnvs().Get(application.Id, env.Name)
		if err != nil {
			return "", "", app.StatusReport{}, err
		}
		provider := providers.Find(workspace.Module.Category, workspace.Module.Type)
		if provider == nil {
			return "", "", app.StatusReport{}, fmt.Errorf("this CLI does not support application category=%s, type=%s", workspace.Module.Category, workspace.Module.Type)
		}
		details := app.Details{
			App:       application,
			Env:       env,
			Workspace: workspace,
		}
		report, err := provider.Status(cfg, details)
		if err != nil {
			return "", "", app.StatusReport{}, fmt.Errorf("error retrieving app status: %w", err)
		}
		version := appEnv.Version
		if version == "" || infraStatus == types.WorkspaceStatusNotProvisioned || infraStatus == "creating" {
			version = "not-deployed"
		}
		return infraStatus, version, report, nil
	}

	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	for {
		buffer := &TableBuffer{}
		buffer.AddFields("env", "infra", "version")
		for _, env := range envs {
			infraStatus, version, report, err := getWorkspaceDetails(env)
			if err != nil {
				return fmt.Errorf("error retrieving app workspace (%s/%s): %w", application.Name, env.Name, err)
			}
			buffer.AddFields(report.Fields...)
			cur := map[string]interface{}{
				"env":     env.Name,
				"infra":   infraStatus,
				"version": version,
			}
			for k, v := range report.Data {
				cur[k] = v
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
	return nil
}
