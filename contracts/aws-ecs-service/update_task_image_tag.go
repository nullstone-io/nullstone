package aws_ecs_service

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
)

func UpdateTaskImageTag(ctx context.Context, infra Outputs, taskDefinition *ecstypes.TaskDefinition, imageTag string) (*ecstypes.TaskDefinition, error) {
	ecsClient := ecs.NewFromConfig(nsaws.NewConfig(infra.Deployer, infra.Region))

	defIndex, err := findMainContainerDefinitionIndex(infra.MainContainerName, taskDefinition.ContainerDefinitions)
	if err != nil {
		return nil, err
	}

	existingImageUrl := docker.ParseImageUrl(*taskDefinition.ContainerDefinitions[defIndex].Image)
	existingImageUrl.Digest = ""
	existingImageUrl.Tag = imageTag
	taskDefinition.ContainerDefinitions[defIndex].Image = aws.String(existingImageUrl.String())

	out, err := ecsClient.RegisterTaskDefinition(ctx, &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions:    taskDefinition.ContainerDefinitions,
		Family:                  taskDefinition.Family,
		Cpu:                     taskDefinition.Cpu,
		ExecutionRoleArn:        taskDefinition.ExecutionRoleArn,
		InferenceAccelerators:   taskDefinition.InferenceAccelerators,
		IpcMode:                 taskDefinition.IpcMode,
		Memory:                  taskDefinition.Memory,
		NetworkMode:             taskDefinition.NetworkMode,
		PidMode:                 taskDefinition.PidMode,
		PlacementConstraints:    taskDefinition.PlacementConstraints,
		ProxyConfiguration:      taskDefinition.ProxyConfiguration,
		RequiresCompatibilities: taskDefinition.RequiresCompatibilities,
		TaskRoleArn:             taskDefinition.TaskRoleArn,
		Volumes:                 taskDefinition.Volumes,
	})
	if err != nil {
		return nil, err
	}

	_, err = ecsClient.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: taskDefinition.TaskDefinitionArn,
	})
	if err != nil {
		return nil, err
	}

	return out.TaskDefinition, nil
}

func findMainContainerDefinitionIndex(mainContainerName string, containerDefs []ecstypes.ContainerDefinition) (int, error) {
	if len(containerDefs) == 0 {
		return -1, fmt.Errorf("cannot deploy service with no container definitions")
	}
	if len(containerDefs) == 1 {
		return 0, nil
	}

	if mainContainerName != "" {
		// let's go find main_container_name
		for i, cd := range containerDefs {
			if cd.Name != nil && *cd.Name == mainContainerName {
				return i, nil
			}
		}
		return -1, fmt.Errorf("cannot deploy service; no container definition with main_container_name = %s", mainContainerName)
	}

	// main_container_name was not specified, we are going to attempt to find a single container definition
	// If more than one container definition exists, we will error
	if len(containerDefs) > 1 {
		return -1, fmt.Errorf("service contains multiple containers; cannot deploy unless service module exports 'main_container_name'")
	}
	return 0, nil
}
