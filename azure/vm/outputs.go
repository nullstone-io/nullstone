package vm

import (
	"github.com/nullstone-io/deployment-sdk/azure"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	SubscriptionId  string          `ns:"subscription_id"`
	ResourceGroup   string          `ns:"resource_group"`
	VmName          string          `ns:"vm_name"`
	BastionHostName string          `ns:"bastion_host_name,optional"`
	Remoter         azure.Principal `ns:"deployer,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	o.Remoter.InitializeCreds(source, ws, nstypes.AutomationPurposeExecRemote, "adminer", "deployer")
}
