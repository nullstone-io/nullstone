package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
)

var Status = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:      "status",
		Usage:     "Application Status",
		UsageText: "nullstone status [options] <app-name> [<env-name>]",
		Flags: []cli.Flag{
			StackFlag,
			AppVersionFlag,
		},
		Action: func(c *cli.Context) error {
			return ProfileAction(c, func(cfg api.Config) error {
				var appName, envName string
				stackName := c.String("stack-name")
				switch c.NArg() {
				case 1:
					appName = c.Args().Get(0)
					// TODO: Support -w/--watch
					return appStatus(cfg, providers, stackName, appName)
				case 2:
					appName = c.Args().Get(0)
					envName = c.Args().Get(1)
					// TODO: Support -w/--watch
					return appEnvStatus(cfg, stackName, appName, envName)
				default:
					cli.ShowCommandHelp(c, c.Command.Name)
					return fmt.Errorf("invalid usage")
				}
			})
		},
	}
}

func appStatus(cfg api.Config, providers app.Providers, stackName, appName string) error {
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
			return "", "", nil, err
		} else if workspace == nil {
			return types.WorkspaceStatusNotProvisioned, "not-deployed", nil, nil
		}
		appEnv, err := client.AppEnvs().Get(application.Id, env.Name)
		if err != nil {
			return "", "", nil, err
		}
		provider := providers.Find(workspace.Module.Category, workspace.Module.Type)
		if provider == nil {
			return "", "", nil, fmt.Errorf("this CLI does not support application category=%s, type=%s", workspace.Module.Category, workspace.Module.Type)
		}
		details := app.Details{
			App:       application,
			Env:       env,
			Workspace: workspace,
		}
		report, err := provider.Status(cfg, details)
		if err != nil {
			return "", "", nil, fmt.Errorf("error retrieving app status: %w", err)
		}
		return workspace.Status, appEnv.Version, report, nil
	}

	buffer := &TableBuffer{}
	buffer.AddFields("env", "infra", "version")
	for _, env := range envs {
		infraStatus, version, report, err := getWorkspaceDetails(env)
		if err != nil {
			return fmt.Errorf("error retrieving app workspace (%s/%s): %w", application.Name, env.Name, err)
		}
		cur := map[string]interface{}{
			"env":     env.Name,
			"infra":   infraStatus,
			"version": version,
		}
		for k, v := range report {
			cur[k] = v
		}
		buffer.AddRow(cur)
	}

	fmt.Println(buffer.Serialize("|"))
	return nil
}

func appEnvStatus(cfg api.Config, stackName, appName, envName string) error {
	return nil
}
