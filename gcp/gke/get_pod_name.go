package gke

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetPodName finds a pod based on the current infrastructure and an optional pod name
// If pod name is left blank, this will find either the only active pod or first active pod in a replica set
func GetPodName(ctx context.Context, cfg *rest.Config, infra Outputs, pod string) (string, error) {
	name := infra.ServiceName
	kubeClient, err := kubernetes.NewForConfig(cfg)

	// If pod is specified, let's verify it exists and is running
	if pod != "" {
		pod, err := kubeClient.CoreV1().Pods(infra.ServiceNamespace).Get(ctx, pod, meta_v1.GetOptions{})
		if err != nil {
			return "", err
		}
		if pod == nil {
			return "", fmt.Errorf("could not find pod (%s) in kubernetes cluster", pod)
		}
		if pod.Status.Phase != v1.PodRunning {
			return "", fmt.Errorf("pod (%s) is not running (current pod status = %s)", pod, pod.Status.Phase)
		}
		return pod.Name, nil
	}

	// If replicas>1, the pods have unique names, but the replicaset has name=<service-name>
	// Let's look for pods by replicaset first
	listOptions := meta_v1.ListOptions{FieldSelector: fmt.Sprintf("replicaset=%s", name)}
	podsOutput, err := kubeClient.CoreV1().Pods(infra.ServiceNamespace).List(ctx, listOptions)
	if err != nil {
		return "", err
	}
	for _, pod := range podsOutput.Items {
		if pod.Status.Phase == v1.PodRunning {
			return pod.Name, nil
		}
	}

	// If we don't find a replica, look for the pod by name directly
	listOptions = meta_v1.ListOptions{LabelSelector: fmt.Sprintf("nullstone.io/app=%s", name)}
	podsOutput, err = kubeClient.CoreV1().Pods(infra.ServiceNamespace).List(ctx, listOptions)
	if err != nil {
		return "", err
	}
	for _, pod := range podsOutput.Items {
		if pod.Status.Phase == v1.PodRunning {
			return pod.Name, nil
		}
	}

	return name, nil
}
