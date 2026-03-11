package appservice

import (
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/azure"
)

type Outputs struct {
	ResourceGroup string          `ns:"resource_group"`
	SiteName      string          `ns:"site_name"`
	Deployer      azure.Principal `ns:"deployer,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	factory := azure.NewTokenProviderFactory(source, ws.StackId, ws.Uid)
	o.Deployer.RemoteTokenProvider = factory("adminer", "deployer")
}
