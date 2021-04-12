package ecs

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func UpdateServiceTask(awsConfig aws.Config, clusterArn string, serviceName string, taskDefinitionArn string) error {
	client := ecs.NewFromConfig(awsConfig)

	_, err := client.UpdateService(context.Background(), &ecs.UpdateServiceInput{
		Service:            aws.String(serviceName),
		Cluster:            aws.String(clusterArn),
		ForceNewDeployment: true,
		TaskDefinition:     aws.String(taskDefinitionArn),
	})
	return err
}
