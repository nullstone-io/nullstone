package gcp_gke

import (
	"context"
	"fmt"
	gcp_gke_service "gopkg.in/nullstone-io/nullstone.v0/contracts/gcp-gke-service"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
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

func (c InfraConfig) ReplacePodSpecImageTag(spec core_v1.PodSpec, imageTag string) (core_v1.PodSpec, error) {
	result := spec

	containerIndex, err := c.findMainContainerDefinitionIndex(result.Containers)
	if err != nil {
		return result, err
	}

	existingImageUrl := docker.ParseImageUrl(result.Containers[containerIndex].Image)
	existingImageUrl.Digest = ""
	existingImageUrl.Tag = imageTag
	result.Containers[containerIndex].Image = existingImageUrl.String()

	return result, nil
}

func (c InfraConfig) findMainContainerDefinitionIndex(containers []core_v1.Container) (int, error) {
	if len(containers) == 0 {
		return -1, fmt.Errorf("cannot deploy service with no containers")
	}
	if len(containers) == 1 {
		return 0, nil
	}

	if mainContainerName := c.Outputs.MainContainerName; mainContainerName != "" {
		// let's go find main_container_name
		for i, container := range containers {
			if container.Name == mainContainerName {
				return i, nil
			}
		}
		return -1, fmt.Errorf("cannot deploy service; no container definition with main_container_name = %s", mainContainerName)
	}

	// main_container_name was not specified, we are going to attempt to find a single container definition
	// If more than one container definition exists, we will error
	if len(containers) > 1 {
		return -1, fmt.Errorf("service contains multiple containers; cannot deploy unless service module exports 'main_container_name'")
	}
	return 0, nil
}

func (c InfraConfig) UpdatePod(pod *core_v1.Pod) (*core_v1.Pod, error) {
	ctx := context.Background()
	conn, err := c.createKubeClient()
	if err != nil {
		return nil, err
	}
	return conn.CoreV1().Pods(c.Outputs.Namespace).Update(ctx, pod, meta_v1.UpdateOptions{})
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
