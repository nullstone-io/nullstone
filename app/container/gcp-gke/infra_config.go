package gcp_gke

import (
	"context"
	"fmt"
	gcp_gke_service "gopkg.in/nullstone-io/nullstone.v0/contracts/gcp-gke-service"
	"gopkg.in/nullstone-io/nullstone.v0/k8s"
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"log"
)

// InfraConfig provides a minimal understanding of the infrastructure provisioned for a module type=aws-fargate
type InfraConfig struct {
	Outputs gcp_gke_service.Outputs

	KubeClient *kubernetes.Clientset
}

func (c *InfraConfig) Print(logger *log.Logger) {
	logger = log.New(logger.Writer(), "    ", 0)
}

func (c *InfraConfig) GetDeployment() (*apps_v1.Deployment, error) {
	ctx := context.TODO()
	conn, err := c.createKubeClient(ctx)
	if err != nil {
		return nil, err
	}
	return conn.AppsV1().Deployments(c.Outputs.ServiceNamespace).Get(ctx, c.Outputs.ServiceName, meta_v1.GetOptions{})
}

func (c *InfraConfig) UpdateDeployment(deployment *apps_v1.Deployment) (*apps_v1.Deployment, error) {
	ctx := context.TODO()
	conn, err := c.createKubeClient(ctx)
	if err != nil {
		return nil, err
	}
	return conn.AppsV1().Deployments(c.Outputs.ServiceNamespace).Update(ctx, deployment, meta_v1.UpdateOptions{})
}

func (c *InfraConfig) GetServices() (*core_v1.ServiceList, error) {
	ctx := context.TODO()
	conn, err := c.createKubeClient(ctx)
	if err != nil {
		return nil, err
	}
	return conn.CoreV1().Services(c.Outputs.ServiceNamespace).List(ctx, meta_v1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", c.Outputs.ServiceName),
	})
}

func (c *InfraConfig) createKubeClient(ctx context.Context) (*kubernetes.Clientset, error) {
	if c.KubeClient == nil {
		clusterOutputs := c.Outputs.Cluster
		configCreator := &k8s.ConfigCreator{
			TokenSourcer:  clusterOutputs.Deployer,
			ClusterInfoer: clusterOutputs,
		}
		cfg, err := configCreator.Create(ctx)
		if err != nil {
			return nil, fmt.Errorf("error creating kube config: %w", err)
		}
		c.KubeClient, err = kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, err
		}
	}
	return c.KubeClient, nil
}
