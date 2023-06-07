package gke

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/gcp/gke"
	"gopkg.in/nullstone-io/nullstone.v0/k8s"
)

func ExecCommand(ctx context.Context, infra Outputs, pod, container string, cmd []string, opts *k8s.ExecOptions) error {
	cfg, err := gke.CreateKubeConfig(ctx, infra.ClusterNamespace, infra.Deployer)
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
