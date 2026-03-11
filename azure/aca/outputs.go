package aca

import (
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/azure"
)

type Outputs struct {
	ResourceGroup     string          `ns:"resource_group"`
	ContainerAppName  string          `ns:"container_app_name,optional"`
	MainContainerName string          `ns:"main_container_name,optional"`
	JobName           string          `ns:"job_name,optional"`
	Deployer          azure.Principal `ns:"deployer,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	factory := azure.NewTokenProviderFactory(source, ws.StackId, ws.Uid)
	o.Deployer.RemoteTokenProvider = factory("adminer", "deployer")
}
