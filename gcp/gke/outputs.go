package gke

import (
	"github.com/nullstone-io/deployment-sdk/gcp"
	"github.com/nullstone-io/deployment-sdk/gcp/creds"
	"github.com/nullstone-io/deployment-sdk/gcp/gke"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

type Outputs struct {
	ServiceNamespace  string             `ns:"service_namespace"`
	ServiceName       string             `ns:"service_name"`
	Runner            gcp.ServiceAccount `ns:"deployer,optional"`
	Remoter           gcp.ServiceAccount `ns:"deployer,optional"`
	MainContainerName string             `ns:"main_container_name,optional"`
	// JobDefinitionName is only specified for a job/task
	// It refers to a Kubernetes ConfigMap containing the job definition in the "template" field
	JobDefinitionName string `ns:"job_definition_name,optional"`

	ClusterNamespace gke.ClusterNamespaceOutputs `ns:",connectionContract:cluster-namespace/gcp/k8s:gke"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *types.Workspace) {
	o.Runner.RemoteTokenSourcer = creds.NewTokenSourcer(source, ws.StackId, ws.BlockId, ws.EnvId, types.AutomationPurposeRun, "deployer")
	o.Remoter.RemoteTokenSourcer = creds.NewTokenSourcer(source, ws.StackId, ws.BlockId, ws.EnvId, types.AutomationPurposeExecRemote, "deployer")
}
