package cmd

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"strings"
)

func newErrMultipleAppsFound(apps []types.Application) *ErrMultipleAppsFound {
	stackNames := make([]string, 0)
	for _, app := range apps {
		stackNames = append(stackNames, app.StackName)
	}
	return &ErrMultipleAppsFound{
		AppName:    apps[0].Name,
		StackNames: stackNames,
	}
}

type ErrMultipleAppsFound struct {
	AppName    string
	StackNames []string
}

func (e ErrMultipleAppsFound) Error() string {
	return fmt.Sprintf("found multiple applications named %q located in the following stacks: %s", e.AppName, strings.Join(e.StackNames, ","))
}

// NsFinder is an object that provides a consistent querying approach for nullstone objects through the CLI
// It provides nice error messages that are tailored for the user flow of CLI commands
type NsFinder struct {
	Config api.Config
}

// FindAppDetails retrieves the app, env, and workspace
// stackName is optional -- If multiple apps are found, this will return an error
func (f NsFinder) FindAppDetails(appName, stackName, envName string) (app.Details, error) {
	appDetails := app.Details{}

	var err error

	if appDetails.App, _, err = f.FindAppAndStack(appName, stackName); err != nil {
		return appDetails, err
	}

	if appDetails.Env, err = f.GetEnv(appDetails.App.StackId, envName); err != nil {
		return appDetails, err
	} else if appDetails.Env == nil {
		return appDetails, fmt.Errorf("environment %s/%s does not exist", stackName, envName)
	}

	if appDetails.Workspace, err = f.getAppWorkspace(appDetails.App, appDetails.Env); err != nil {
		return appDetails, err
	}

	if appDetails.Module, err = find.BlockModule(f.Config, appDetails.App.Block); err != nil {
		return appDetails, err
	} else if appDetails.Module == nil {
		return appDetails, fmt.Errorf("can't find app module %s", appDetails.App.ModuleSource)
	}

	return appDetails, nil
}

func (f NsFinder) FindAppAndStack(appName, stackName string) (*types.Application, *types.Stack, error) {
	var stackId int64
	var stack *types.Stack
	if stackName != "" {
		var err error
		if stack, err = f.FindStack(stackName); err != nil {
			return nil, nil, err
		} else if stack == nil {
			return nil, nil, fmt.Errorf("stack %s does not exist", stackName)
		}
	}

	app, err := f.getApp(appName, stackId)
	if err != nil {
		return nil, nil, err
	}

	return app, stack, nil
}

func (f NsFinder) FindStack(stackName string) (*types.Stack, error) {
	client := api.Client{Config: f.Config}
	stacks, err := client.Stacks().List()
	if err != nil {
		return nil, fmt.Errorf("error retrieving stacks: %w", err)
	}
	for _, stack := range stacks {
		if stack.Name == stackName {
			return stack, nil
		}
	}
	return nil, nil
}

// getApp searches for an app by app name and optionally stack name
// If only 1 app is found, returns that app
// If many are found, will return an error with matched app stack names
func (f NsFinder) getApp(appName string, stackId int64) (*types.Application, error) {
	client := api.Client{Config: f.Config}
	allApps, err := client.Apps().List()
	if err != nil {
		return nil, fmt.Errorf("error listing applications: %w", err)
	}

	matched := make([]types.Application, 0)
	for _, app := range allApps {
		if app.Name == appName && (stackId == 0 || app.StackId == stackId) {
			matched = append(matched, app)
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("application %q does not exist", appName)
	} else if len(matched) > 1 {
		return nil, newErrMultipleAppsFound(matched)
	}

	return &matched[0], nil
}

func (f NsFinder) GetEnv(stackId int64, envName string) (*types.Environment, error) {
	client := api.Client{Config: f.Config}
	envs, err := client.Environments().List(stackId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving environments: %w", err)
	}
	for _, env := range envs {
		if env.Name == envName {
			return env, nil
		}
	}
	return nil, nil
}

func (f NsFinder) getAppWorkspace(app *types.Application, env *types.Environment) (*types.Workspace, error) {
	client := api.Client{Config: f.Config}

	workspace, err := client.Workspaces().Get(app.StackId, app.Id, env.Id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving workspace: %w", err)
	} else if workspace == nil {
		return nil, fmt.Errorf("workspace %q does not exist", err)
	}
	if workspace.Status != types.WorkspaceStatusProvisioned {
		return nil, fmt.Errorf("app %q has not been provisioned in %q environment yet", app.Name, env.Name)
	}
	return workspace, nil
}
