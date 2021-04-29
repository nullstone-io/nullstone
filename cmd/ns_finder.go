package cmd

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
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

// This retrieves the app and workspace
// stackName is optional -- If multiple apps are found, this will return an error
func (f NsFinder) GetAppAndWorkspace(appName, stackName, envName string) (*types.Application, *types.Workspace, error) {
	app, err := f.GetApp(appName, stackName)
	if err != nil {
		return nil, nil, err
	}

	workspace, err := f.GetWorkspace(app.StackName, app.Name, envName)
	if err != nil {
		return nil, nil, err
	}
	return app, workspace, nil
}

// GetApp searches for an app by app name and optionally stack name
// If only 1 app is found, returns that app
// If many are found, will return an error with matched app stack names
func (f NsFinder) GetApp(appName string, stackName string) (*types.Application, error) {
	client := api.Client{Config: f.Config}
	allApps, err := client.Apps().List()
	if err != nil {
		return nil, fmt.Errorf("error listing applications: %w", err)
	}

	matched := make([]types.Application, 0)
	for _, app := range allApps {
		if app.Name == appName && (stackName == "" || app.StackName == stackName) {
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

func (f NsFinder) GetWorkspace(stackName, blockName, envName string) (*types.Workspace, error) {
	client := api.Client{Config: f.Config}
	workspace, err := client.Workspaces().Get(stackName, blockName, envName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving workspace: %w", err)
	} else if workspace == nil {
		return nil, fmt.Errorf("workspace %q does not exist", err)
	}
	if workspace.Status != types.WorkspaceStatusProvisioned {
		return nil, fmt.Errorf("app %q has not been provisioned in %q environment yet", blockName, workspace.EnvName)
	}
	if workspace.Module == nil {
		return nil, fmt.Errorf("unknown module for workspace")
	}
	return workspace, nil
}