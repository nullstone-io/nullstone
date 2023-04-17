package gke

import (
	"context"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetPodName(ctx context.Context, cfg *rest.Config, infra Outputs) (string, error) {
	podName := infra.ServiceName

	kubeClient, err := kubernetes.NewForConfig(cfg)
	deployment, err := kubeClient.AppsV1().Deployments(infra.ServiceNamespace).Get(ctx, infra.ServiceName, meta_v1.GetOptions{})
	if err != nil {
		return "", err
	}


	replicaSets, err := kubeClient.AppsV1().ReplicaSets(infra.ServiceNamespace).Get(ctx, infra.ServiceNamespace, meta_v1.GetOptions{})

	replicaSets.Status.
}
