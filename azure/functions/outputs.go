package functions

import (
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/azure"
)

type Outputs struct {
	Deployer        azure.Principal `ns:"deployer,optional"`
	FunctionAppName string          `ns:"function_app_name"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	factory := azure.NewTokenProviderFactory(source, ws.StackId, ws.Uid)
	o.Deployer.RemoteTokenProvider = factory("deployer")
}
