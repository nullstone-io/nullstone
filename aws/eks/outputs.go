package eks

import (
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"github.com/nullstone-io/deployment-sdk/aws/creds"
	"github.com/nullstone-io/deployment-sdk/aws/eks"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	ServiceNamespace  string     `ns:"service_namespace"`
	ServiceName       string     `ns:"service_name"`
	Deployer          nsaws.User `ns:"deployer,optional"`
	MainContainerName string     `ns:"main_container_name,optional"`
	// JobDefinitionName is only specified for a job/task
	// It refers to a Kubernetes ConfigMap containing the job definition in the "template" field
	JobDefinitionName string `ns:"job_definition_name,optional"`

	ClusterNamespace eks.ClusterNamespaceOutputs `ns:",connectionContract:cluster-namespace/aws/k8s:eks"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	credsFactory := creds.NewProviderFactory(source, ws.StackId, ws.Uid)
	o.Deployer.RemoteProvider = credsFactory("deployer")
}
