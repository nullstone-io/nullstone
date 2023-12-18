package ecs

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/nullstone-io/deployment-sdk/app"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"golang.org/x/sync/errgroup"
	"log"
	"strings"
	"time"
)

const DefaultWatchInterval = 1 * time.Second

func RunTask(ctx context.Context, infra Outputs, containerName, username string, cmd []string, logStreamer app.LogStreamer, logEmitter app.LogEmitter) error {
	region := infra.Region
	awsConfig := nsaws.NewConfig(infra.Deployer, region)
	ecsClient := ecs.NewFromConfig(awsConfig)

	latestArn, err := getTaskDefArn(ctx, ecsClient, infra.TaskArn)
	if err != nil {
		return fmt.Errorf("unable to find the latest task definition arn: %w", err)
	}
	infra.TaskArn = latestArn

	out, err := ecsClient.RunTask(ctx, createTaskInput(infra, containerName, cmd, username))
	if err != nil {
		return fmt.Errorf("error starting job: %w", err)
	}

	taskArn, err := parseRunTaskResult(out)
	if err != nil {
		return err
	}

	return monitorTask(ctx, logStreamer, logEmitter, ecsClient, infra.ClusterArn(), taskArn, infra.MainContainerName)
}

func getTaskDefArn(ctx context.Context, ecsClient *ecs.Client, taskDefArn string) (string, error) {
	// there might be a newer version of the task definition
	// so we extract the family in order to find the latest task definition arn
	arnParts := strings.Split(taskDefArn, "/")
	familyAndVersion := arnParts[len(arnParts)-1]
	parts := strings.Split(familyAndVersion, ":")
	templateTaskDefName := strings.Join(parts[:len(parts)-1], ":")

	out, err := ecsClient.ListTaskDefinitions(ctx, &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: aws.String(templateTaskDefName),
		Sort:         types.SortOrderDesc,
		Status:       types.TaskDefinitionStatusActive,
	})
	if err != nil {
		return "", err
	}
	for _, tda := range out.TaskDefinitionArns {
		return tda, nil
	}
	return "", nil
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

func getTaskExitCode(ctx context.Context, ecsClient *ecs.Client, clusterArn, taskArn, mainContainerName string) (*int32, error) {
	input := ecs.DescribeTasksInput{
		Tasks:   []string{taskArn},
		Cluster: &clusterArn,
	}
	result, err := ecsClient.DescribeTasks(ctx, &input)
	if err != nil {
		return nil, err
	}
	if len(result.Tasks) == 0 {
		return nil, fmt.Errorf("unable to determine the status of the running task, no tasks found")
	}
	if len(result.Tasks) > 1 {
		return nil, fmt.Errorf("unable to determine the status of the running task, more than one task found")
	}
	if result.Tasks[0].LastStatus == nil {
		return nil, fmt.Errorf("unable to determine the status of the running task, no status returned")
	}
	status := *result.Tasks[0].LastStatus

	var exitCode *int32
	foundContainer := false
	for _, container := range result.Tasks[0].Containers {
		if container.Name != nil && *container.Name == mainContainerName {
			if container.ExitCode != nil {
				exitCode = container.ExitCode
			}
			foundContainer = true
			break
		}
	}
	if !foundContainer {
		return nil, fmt.Errorf("unable to determine the status of the running task, the primary container (%s) could not be found", mainContainerName)
	}

	switch status {
	case "STOPPED":
		return exitCode, nil
	case "DELETED":
		return exitCode, nil
	}

	return nil, nil
}

func monitorTask(ctx context.Context, logStreamer app.LogStreamer, logEmitter app.LogEmitter, ecsClient *ecs.Client, clusterArn, taskArn, mainContainerName string) error {
	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)

	eg.Go(func() error {
		absoluteTime := time.Now()
		logStreamOptions := app.LogStreamOptions{
			StartTime:     &absoluteTime,
			WatchInterval: time.Duration(0), // this makes sure the log stream doesn't exit until the context is cancelled
			Emitter:       logEmitter,
		}
		return logStreamer.Stream(ctx, logStreamOptions)
	})

	eg.Go(func() error {
		for {
			// check status of task
			exitCode, err := getTaskExitCode(ctx, ecsClient, clusterArn, taskArn, mainContainerName)
			if err != nil {
				return err
			}
			if exitCode != nil {
				if *exitCode == 0 {
					log.Printf("Task has completed successfully")
					cancel()
					return nil
				} else {
					return fmt.Errorf("Task failed with status code %d\n", exitCode)
				}
			}

			select {
			case <-ctx.Done():
				switch err := ctx.Err(); {
				case errors.Is(err, context.Canceled):
					return fmt.Errorf("cancelled")
				case errors.Is(err, context.DeadlineExceeded):
					return fmt.Errorf("timeout")
				}
			case <-time.After(DefaultWatchInterval):
			}
		}
	})

	return eg.Wait()
}
