package ecs

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
)

func GetTasks(ctx context.Context, infra Outputs) ([]string, error) {
	ecsClient := ecs.NewFromConfig(nsaws.NewConfig(infra.Deployer, infra.Region))
	out, err := ecsClient.ListTasks(ctx, &ecs.ListTasksInput{
		Cluster:     aws.String(infra.ClusterArn()),
		ServiceName: aws.String(infra.ServiceName),
	})
	if err != nil {
		return nil, err
	}
	return out.TaskArns, nil
}

func GetRandomTask(ctx context.Context, infra Outputs) (string, error) {
	taskArns, err := GetTasks(ctx, infra)
	if err != nil {
		return "", err
	}

	for _, taskArn := range taskArns {
		return taskArn, nil
	}
	return "", nil
}
