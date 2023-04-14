package ecs

import (
	"context"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ssm"
	"strings"
)

func ExecCommand(ctx context.Context, infra Outputs, taskId string, cmd []string, parameters map[string][]string) error {
	region := infra.Region
	cluster := infra.ClusterArn()
	if containerName == "" {
		containerName = infra.MainContainerName
	}
	awsConfig := nsaws.NewConfig(infra.Deployer, region)

	return ssm.StartEcsSession(ctx, awsConfig, region, cluster, taskId, containerName, strings.Join(cmd, " "), parameters)
}
