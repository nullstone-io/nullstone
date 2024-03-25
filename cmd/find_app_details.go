package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// FindAppDetails retrieves the app, env, and workspace
// stackName is optional -- If multiple apps are found, this will return an error
func FindAppDetails(cfg api.Config, appName, stackName, envName string) (app.Details, error) {
	ctx := context.TODO()
	appDetails := app.Details{}

	_, application, env, err := find.StackAppEnv(ctx, cfg, stackName, appName, envName)
	if err != nil {
		return appDetails, err
	}
	appDetails.App = application
	appDetails.Env = env

	if appDetails.Workspace, err = getAppWorkspace(cfg, appDetails.App, appDetails.Env); err != nil {
		return appDetails, err
	}

	if appDetails.Module, err = find.Module(ctx, cfg, appDetails.App.ModuleSource); err != nil {
		return appDetails, err
	} else if appDetails.Module == nil {
		return appDetails, fmt.Errorf("can't find app module %s", appDetails.App.ModuleSource)
	}

	return appDetails, nil
}

func getAppWorkspace(cfg api.Config, app *types.Application, env *types.Environment) (*types.Workspace, error) {
	ctx := context.TODO()
	client := api.Client{Config: cfg}

	workspace, err := client.Workspaces().Get(ctx, app.StackId, app.Id, env.Id)
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
