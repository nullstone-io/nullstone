package functions

import (
	"github.com/nullstone-io/deployment-sdk/azure"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	FunctionAppName string          `ns:"function_app_name"`
	Runner          azure.Principal `ns:"deployer,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	o.Runner.InitializeCreds(source, ws, nstypes.AutomationPurposeRun, "adminer", "deployer")
}
