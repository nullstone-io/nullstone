package eks

import (
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"github.com/nullstone-io/deployment-sdk/aws/eks"
	"github.com/nullstone-io/deployment-sdk/docker"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Outputs struct {
	ServiceNamespace  string          `ns:"service_namespace"`
	ServiceName       string          `ns:"service_name"`
	ImageRepoUrl      docker.ImageUrl `ns:"image_repo_url,optional"`
	Deployer          nsaws.User      `ns:"deployer"`
	MainContainerName string          `ns:"main_container_name,optional"`

	Region           string                  `ns:"region"`
	ClusterNamespace ClusterNamespaceOutputs `ns:",connectionContract:cluster-namespace/aws/k8s:eks"`
}

type ClusterNamespaceOutputs struct {
	ClusterName          string `ns:"cluster_name"`
	ClusterEndpoint      string `ns:"cluster_endpoint"`
	ClusterCACertificate string `ns:"cluster_ca_certificate"`
}

func (o ClusterNamespaceOutputs) GetClusterName() string {
	return o.ClusterName
}

func (o ClusterNamespaceOutputs) ClusterInfo() (clientcmdapi.Cluster, error) {
	return eks.GetClusterInfo(o.ClusterEndpoint, o.ClusterCACertificate)
}
