package ecs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
)

func UpdateTaskImageTag(awsConfig aws.Config, taskDefinition *types.TaskDefinition, imageTag string) (*types.TaskDefinition, error) {
	client := ecs.NewFromConfig(awsConfig)

	defIndex, err := findMainContainerDefinitionIndex(taskDefinition.ContainerDefinitions)
	if err != nil {
		return nil, err
	}

	existingImageUrl := docker.ParseImageUrl(*taskDefinition.ContainerDefinitions[defIndex].Image)
	existingImageUrl.Digest = ""
	existingImageUrl.Tag = imageTag
	taskDefinition.ContainerDefinitions[defIndex].Image = aws.String(existingImageUrl.String())

	out, err := client.RegisterTaskDefinition(context.Background(), &ecs.RegisterTaskDefinitionInput{
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

	_, err = client.DeregisterTaskDefinition(context.Background(), &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: taskDefinition.TaskDefinitionArn,
	})
	if err != nil {
		return nil, err
	}

	return out.TaskDefinition, nil
}

func findMainContainerDefinitionIndex(containerDefs []types.ContainerDefinition) (int, error) {
	mainIndex := -1
	for i, cd := range containerDefs {
		if cd.Essential != nil && *cd.Essential {
			if mainIndex > -1 {
				return 0, fmt.Errorf("cannot deploy a service with multiple containers marked as essential")
			}
			mainIndex = i
		}
	}
	if mainIndex > -1 {
		return mainIndex, nil
	}

	if len(containerDefs) == 0 {
		return 0, fmt.Errorf("cannot deploy service with no container definitions")
	}
	if len(containerDefs) > 1 {
		return 0, fmt.Errorf("cannot deploy service with multiple container definitions unless a single is marked essential")
	}
	return 0, nil
}
