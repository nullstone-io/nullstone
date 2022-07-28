package ecs

import (
	"context"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ssm"
)

func ExecCommand(ctx context.Context, infra Outputs, taskId string, cmd string, parameters map[string][]string) error {
	region := infra.Region
	cluster := infra.Cluster.ClusterArn
	containerName := infra.MainContainerName
	awsConfig := nsaws.NewConfig(infra.Deployer, region)

	return ssm.StartEcsSession(ctx, awsConfig, region, cluster, taskId, containerName, cmd, parameters)
}
