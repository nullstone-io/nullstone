package gke

import (
	"github.com/nullstone-io/deployment-sdk/docker"
	"github.com/nullstone-io/deployment-sdk/gcp"
	"github.com/nullstone-io/deployment-sdk/gcp/gke"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Outputs struct {
	ServiceNamespace  string             `ns:"service_namespace"`
	ServiceName       string             `ns:"service_name"`
	ImageRepoUrl      docker.ImageUrl    `ns:"image_repo_url,optional"`
	Deployer          gcp.ServiceAccount `ns:"deployer"`
	MainContainerName string             `ns:"main_container_name,optional"`

	ClusterNamespace ClusterNamespaceOutputs `ns:",connectionContract:cluster-namespace/gcp/k8s:gke"`
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
