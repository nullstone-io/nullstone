package ecs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"log"
)

func RunTask(ctx context.Context, infra Outputs, containerName, username string, cmd []string) error {
	log.Printf("Infra outputs: %#v\n", infra)
	log.Printf("Command: %s\n", cmd)

	region := infra.Region
	awsConfig := nsaws.NewConfig(infra.Deployer, region)
	ecsClient := ecs.NewFromConfig(awsConfig)

	out, err := ecsClient.RunTask(ctx, createTaskInput(infra, containerName, cmd, username))
	if err != nil {
		return fmt.Errorf("error starting job: %w", err)
	}

	_, err = parseRunTaskResult(out)
	if err != nil {
		return err
	}
	return nil
}

func createTaskInput(infra Outputs, containerName string, cmd []string, createdBy string) *ecs.RunTaskInput {
	clusterArn := infra.ClusterArn()
	subnetIds := infra.PrivateSubnetIds()
	securityGroupIds := []string{infra.AppSecurityGroupId}
	if containerName == "" {
		containerName = infra.MainContainerName
	}
	taskDefArn := infra.TaskArn
	launchType := infra.GetLaunchType()

	log.Printf("Cluster Arn: %s\n", clusterArn)

	return &ecs.RunTaskInput{
		TaskDefinition:       aws.String(taskDefArn),
		Cluster:              aws.String(clusterArn),
		Count:                aws.Int32(1),
		LaunchType:           launchType,
		EnableECSManagedTags: true,
		EnableExecuteCommand: false,
		NetworkConfiguration: &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				Subnets:        subnetIds,
				AssignPublicIp: types.AssignPublicIpDisabled,
				SecurityGroups: securityGroupIds,
			},
		},
		Overrides: &types.TaskOverride{
			ContainerOverrides: []types.ContainerOverride{
				{
					Name:    aws.String(containerName),
					Command: cmd,
				},
			},
		},
		PropagateTags: types.PropagateTagsTaskDefinition,
		StartedBy:     aws.String(createdBy),
	}
}

func parseRunTaskResult(out *ecs.RunTaskOutput) (string, error) {
	for _, failure := range out.Failures {
		return "", fmt.Errorf("builder failed to start: %v", failure)
	}

	for _, task := range out.Tasks {
		if task.TaskArn != nil {
			return *task.TaskArn, nil
		}
	}

	return "", nil
}
