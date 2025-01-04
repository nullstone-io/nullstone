package gke

import (
	"github.com/nullstone-io/deployment-sdk/docker"
	"github.com/nullstone-io/deployment-sdk/gcp"
	"github.com/nullstone-io/deployment-sdk/gcp/creds"
	"github.com/nullstone-io/deployment-sdk/gcp/gke"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Outputs struct {
	ServiceNamespace  string             `ns:"service_namespace"`
	ServiceName       string             `ns:"service_name"`
	ImageRepoUrl      docker.ImageUrl    `ns:"image_repo_url,optional"`
	Deployer          gcp.ServiceAccount `ns:"deployer"`
	MainContainerName string             `ns:"main_container_name,optional"`
	// JobDefinition is only specified for a job/task
	// It is a base64-encoded JSON string that contains the manifest for a job
	JobDefinition string `ns:"job_definition,optional"`

	ClusterNamespace ClusterNamespaceOutputs `ns:",connectionContract:cluster-namespace/gcp/k8s:gke"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	o.Deployer.RemoteTokenSourcer = creds.NewTokenSourcer(source, ws.StackId, ws.Uid, "deployer")
}

type ClusterNamespaceOutputs struct {
	ClusterEndpoint      string `ns:"cluster_endpoint"`
	ClusterCACertificate string `ns:"cluster_ca_certificate"`
}

func (o ClusterNamespaceOutputs) ClusterInfo() (clientcmdapi.Cluster, error) {
	return gke.GetClusterInfo(o.ClusterEndpoint, o.ClusterCACertificate)
}

type ClusterOutputs struct {
	ClusterEndpoint      string `ns:"cluster_endpoint"`
	ClusterCACertificate string `ns:"cluster_ca_certificate"`
}

func (o ClusterOutputs) ClusterInfo() (clientcmdapi.Cluster, error) {
	return gke.GetClusterInfo(o.ClusterEndpoint, o.ClusterCACertificate)
}
