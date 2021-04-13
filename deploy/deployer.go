package deploy

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"log"
)

type Deployer interface {
	// Detect allows the Deployer to determine whether this app is compatible to deploy through this Deployer
	Detect(app *types.Application, workspace *types.Workspace) bool

	// Identify collects necessary information from the workspace (and connected workspaces) needed to run deployment
	// Typically, this will identify the app.Category and workspace.Module.Type (module type usually contains provider)
	Identify(nsConfig api.Config, app *types.Application, workspace *types.Workspace) (InfraConfig, error)

	// Deploy performs the necessary steps to deploy the application
	Deploy(app *types.Application, workspace *types.Workspace, userConfig map[string]string, infraConfig interface{}) error
}

type InfraConfig interface {
	Print(logger *log.Logger)
}
