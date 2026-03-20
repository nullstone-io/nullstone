package aca

import (
	"github.com/nullstone-io/deployment-sdk/azure"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	ResourceGroup     string          `ns:"resource_group"`
	ContainerAppName  string          `ns:"container_app_name,optional"`
	MainContainerName string          `ns:"main_container_name,optional"`
	JobName           string          `ns:"job_name,optional"`
	Runner            azure.Principal `ns:"deployer,optional"`
	Remoter           azure.Principal `ns:"deployer,optional"`
	Statuser          azure.Principal `ns:"deployer,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	o.Runner.InitializeCreds(source, ws, nstypes.AutomationPurposeRun, "adminer", "deployer")
	o.Remoter.InitializeCreds(source, ws, nstypes.AutomationPurposeExecRemote, "adminer", "deployer")
	o.Statuser.InitializeCreds(source, ws, nstypes.AutomationPurposeViewStatus, "adminer", "deployer")
}
