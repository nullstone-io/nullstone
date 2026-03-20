package aks

import (
	"encoding/base64"
	"fmt"

	"github.com/nullstone-io/deployment-sdk/azure"
	"github.com/nullstone-io/deployment-sdk/k8s"
	"github.com/nullstone-io/deployment-sdk/outputs"
	nstypes "gopkg.in/nullstone-io/go-api-client.v0/types"
	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Outputs struct {
	ServiceNamespace  string          `ns:"service_namespace"`
	ServiceName       string          `ns:"service_name"`
	Runner            azure.Principal `ns:"deployer,optional"`
	Remoter           azure.Principal `ns:"deployer,optional"`
	MainContainerName string          `ns:"main_container_name,optional"`
	// JobDefinitionName is only specified for a job/task
	// It refers to a Kubernetes ConfigMap containing the job definition in the "template" field
	JobDefinitionName string `ns:"job_definition_name,optional"`

	ClusterNamespace ClusterNamespaceOutputs `ns:",connectionContract:cluster-namespace/azure/k8s:aks"`
}

func (o *Outputs) InitializeCreds(source outputs.RetrieverSource, ws *nstypes.Workspace) {
	o.Runner.InitializeCreds(source, ws, nstypes.AutomationPurposeRun, "adminer", "deployer")
	o.Remoter.InitializeCreds(source, ws, nstypes.AutomationPurposeExecRemote, "adminer", "deployer")
}

type ClusterNamespaceOutputs struct {
	ClusterEndpoint      string `ns:"cluster_endpoint"`
	ClusterCACertificate string `ns:"cluster_ca_certificate"`
}

var _ k8s.ClusterInfoer = ClusterNamespaceOutputs{}

func (o ClusterNamespaceOutputs) ClusterInfo() (clientcmdapi.Cluster, error) {
	return GetClusterInfo(o.ClusterEndpoint, o.ClusterCACertificate)
}

func GetClusterInfo(endpoint string, caCertificate string) (clientcmdapi.Cluster, error) {
	decodedCACert, err := base64.StdEncoding.DecodeString(caCertificate)
	if err != nil {
		return clientcmdapi.Cluster{}, fmt.Errorf("invalid cluster CA certificate: %w", err)
	}

	host, _, err := restclient.DefaultServerURL(endpoint, "", apimachineryschema.GroupVersion{Group: "", Version: "v1"}, true)
	if err != nil {
		return clientcmdapi.Cluster{}, fmt.Errorf("failed to parse AKS cluster host %q: %w", endpoint, err)
	}

	return clientcmdapi.Cluster{
		Server:                   host.String(),
		CertificateAuthorityData: decodedCACert,
	}, nil
}
