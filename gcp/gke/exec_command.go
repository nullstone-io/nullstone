package gke

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/nullstone.v0/k8s"
)

func ExecCommand(ctx context.Context, infra Outputs, task string, cmd []string, opts *k8s.ExecOptions) error {
	cfg, err := CreateRESTConfig(ctx, infra.Deployer, infra.ClusterNamespace)
	if err != nil {
		return fmt.Errorf("error creating kube config: %w", err)
	}

	podName, err := GetPodName(ctx, cfg, infra)

	// TODO: The service name may refer to a replica set -- we need to use `task` to identify the specific pod in the replica set
	// TODO: Find container name if blank
	// TODO: Verify pod is not corev1.PodSucceeded or corev1.PodFailed
	//if task == "" {
	//	var err error
	//	if task, err = GetRandomTask(ctx, r.Infra); err != nil {
	//		return err
	//	} else if task == "" {
	//		return fmt.Errorf("cannot exec command with no running tasks")
	//	}
	//}

	return k8s.ExecCommand(ctx, cfg, infra.ServiceNamespace, podName, infra.MainContainerName, cmd, opts)
}
