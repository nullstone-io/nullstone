package lambda

import (
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"github.com/nullstone-io/deployment-sdk/aws/creds"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	Region     string            `ns:"region"`
	LambdaArn  string            `ns:"lambda_arn"`
	LambdaName string            `ns:"lambda_name"`
	Runner     nsaws.IamIdentity `ns:"deployer,optional"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	credsFactory := creds.NewProviderFactory(source, ws.StackId, ws.BlockId, ws.EnvId)
	o.Runner.RemoteProvider = credsFactory(nstypes.AutomationPurposeRun, "deployer")
}
