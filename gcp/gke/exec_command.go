package gke

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/nullstone.v0/k8s"
)

func ExecCommand(ctx context.Context, infra Outputs, pod, container string, cmd []string, opts *k8s.ExecOptions) error {
	cfg, err := CreateRESTConfig(ctx, infra.Deployer, infra.ClusterNamespace)
	if err != nil {
		return fmt.Errorf("error creating kube config: %w", err)
	}

	podName, err := GetPodName(ctx, cfg, infra, pod)
	if err != nil {
		return fmt.Errorf("error finding pod: %w", err)
	}

	if container == "" {
		container = infra.MainContainerName
	}

	return k8s.ExecCommand(ctx, cfg, infra.ServiceNamespace, podName, container, cmd, opts)
}
