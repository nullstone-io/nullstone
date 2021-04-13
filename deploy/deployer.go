package deploy

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"log"
)

type Deployer interface {
	Detect(app *types.Application, workspace *types.Workspace) bool
	Identify(nsConfig api.Config, app *types.Application, workspace *types.Workspace) (InfraConfig, error)
	Deploy(app *types.Application, workspace *types.Workspace, userConfig map[string]string, infraConfig interface{}) error
}

type InfraConfig interface {
	Print(logger *log.Logger)
}
