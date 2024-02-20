package cmd

import (
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
)

// FindAppDetails retrieves the app, env, and workspace
// stackName is optional -- If multiple apps are found, this will return an error
func FindAppDetails(cfg api.Config, appName, stackName, envName string) (app.Details, error) {
	appDetails := app.Details{}

	_, application, env, err := find.StackAppEnv(cfg, stackName, appName, envName)
	if err != nil {
		return appDetails, err
	}
	appDetails.App = application
	appDetails.Env = env

	client := api.Client{Config: cfg}
	appDetails.Workspace, err = client.Workspaces().Get(appDetails.App.StackId, appDetails.App.Id, env.Id)
	if err != nil {
		return appDetails, fmt.Errorf("error retrieving workspace: %w", err)
	} else if appDetails.Workspace == nil {
		return appDetails, fmt.Errorf("workspace %q does not exist", err)
	}

	if appDetails.Module, err = find.Module(cfg, appDetails.App.ModuleSource); err != nil {
		return appDetails, err
	} else if appDetails.Module == nil {
		return appDetails, fmt.Errorf("can't find app module %s", appDetails.App.ModuleSource)
	}

	return appDetails, nil
}
