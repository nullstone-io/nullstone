package cloudrun

import (
	"github.com/nullstone-io/deployment-sdk/gcp"
	"github.com/nullstone-io/deployment-sdk/gcp/creds"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	ServiceName       string             `ns:"service_name"`
	Runner            gcp.ServiceAccount `ns:"deployer,optional"`
	Remoter           gcp.ServiceAccount `ns:"deployer,optional"`
	MainContainerName string             `ns:"main_container_name,optional"`
	JobId             string             `ns:"job_id,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	o.Runner.RemoteTokenSourcer = creds.NewTokenSourcer(source, ws.StackId, ws.BlockId, ws.EnvId, nstypes.AutomationPurposeRun, "deployer")
	o.Remoter.RemoteTokenSourcer = creds.NewTokenSourcer(source, ws.StackId, ws.BlockId, ws.EnvId, nstypes.AutomationPurposeExecRemote, "deployer")
}
