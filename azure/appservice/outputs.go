package appservice

import (
	"github.com/nullstone-io/deployment-sdk/azure"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	ResourceGroup string          `ns:"resource_group"`
	SiteName      string          `ns:"site_name"`
	Remoter       azure.Principal `ns:"deployer,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	o.Remoter.InitializeCreds(source, ws, nstypes.AutomationPurposeExecRemote, "adminer", "deployer")
}
