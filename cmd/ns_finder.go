package cmd

import (
	"errors"
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

// This retrieves the app, env, and workspace
// stackName is optional -- If multiple apps are found, this will return an error
func (f NsFinder) GetAppAndWorkspace(appName, stackName, envName string) (*types.Application, *types.Environment, *types.Workspace, error) {
	var stackId int64
	if stackName != "" {
		stack, err := f.GetStack(stackName)
		if err != nil {
			return nil, nil, nil, err
		} else if stack == nil {
			return nil, nil, nil, fmt.Errorf("stack %s does not exist", stackName)
		}
	}

	app, err := f.GetApp(appName, stackId)
	if err != nil {
		return nil, nil, nil, err
	}

	env, err := f.GetEnv(app.StackId, envName)
	if err != nil {
		return nil, nil, nil, err
	} else if env == nil {
		return nil, nil, nil, fmt.Errorf("environment %s/%s does not exist", stackName, envName)
	}

	workspace, err := f.GetAppWorkspace(app, env)
	if err != nil {
		return nil, nil, nil, err
	}

	return app, env, workspace, nil
}

// GetApp searches for an app by app name and optionally stack name
// If only 1 app is found, returns that app
// If many are found, will return an error with matched app stack names
func (f NsFinder) GetApp(appName string, stackId int64) (*types.Application, error) {
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

func (f NsFinder) GetStack(stackName string) (*types.Stack, error) {
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

func (f NsFinder) GetAppWorkspace(app *types.Application, env *types.Environment) (*types.Workspace, error) {
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
	if workspace.Module == nil {
		return nil, fmt.Errorf("unknown module for workspace")
	}
	return workspace, nil
}

func (f NsFinder) GetAppModule(client api.Client, app types.Application) (*types.Module, error) {
	ms, err := ParseSource(app.ModuleSource)
	if err != nil {
		return nil, err
	}
	return client.Org(ms.OrgName).Modules().Get(ms.ModuleName)
}

var ErrInvalidModuleSource = errors.New("invalid module source")

type ModuleSource struct {
	Host       string
	OrgName    string
	ModuleName string
}

func ParseSource(source string) (*ModuleSource, error) {
	tokens := strings.Split(source, "/")
	switch len(tokens) {
	case 2:
		// nullstone registry implied
		return &ModuleSource{
			Host:       "",
			OrgName:    tokens[0],
			ModuleName: tokens[1],
		}, nil
	case 3:
		return &ModuleSource{
			Host:       tokens[0],
			OrgName:    tokens[1],
			ModuleName: tokens[2],
		}, nil
	default:
		// this does not match anything resembling a nullstone registry source
		return nil, ErrInvalidModuleSource
	}
}
