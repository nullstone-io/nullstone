package gcp_gke

import (
	"context"
	"fmt"
	gcp_gke_service "gopkg.in/nullstone-io/nullstone.v0/contracts/gcp-gke-service"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

// InfraConfig provides a minimal understanding of the infrastructure provisioned for a module type=aws-fargate
type InfraConfig struct {
	Outputs gcp_gke_service.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger = log.New(logger.Writer(), "    ", 0)
}

func (c InfraConfig) GetPod() (*core_v1.Pod, error) {
	ctx := context.Background()
	conn, err := c.createKubeClient()
	if err != nil {
		return nil, err
	}
	return conn.CoreV1().Pods(c.Outputs.Namespace).Get(ctx, c.Outputs.Name, meta_v1.GetOptions{})
}

func (c InfraConfig) createKubeClient() (*kubernetes.Clientset, error) {
	cfg, err := c.createKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid GKE configuration: %w", err)
	}
	return kubernetes.NewForConfig(cfg)
}

func (c InfraConfig) createKubeConfig() (*restclient.Config, error) {
	clusterOutputs := c.Outputs.Cluster

	overrides := &clientcmd.ConfigOverrides{}
	loader := &clientcmd.ClientConfigLoadingRules{}

	overrides.ClusterInfo.CertificateAuthorityData = []byte(c.Outputs.Cluster.ClusterCACertificate)
	host, _, err := restclient.DefaultServerURL(clusterOutputs.ClusterEndpoint, "", apimachineryschema.GroupVersion{}, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GKE cluster host %q: %w", clusterOutputs.ClusterEndpoint, err)
	}
	overrides.ClusterInfo.Server = host.String()

	overrides.AuthInfo.TokenFile = clusterOutputs.Deployer.Token

	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides)
	return cc.ClientConfig()
}
