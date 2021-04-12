package ecs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func GetTaskDefinitionByServiceInCluster(awsConfig aws.Config, clusterArn, serviceName string) (*types.TaskDefinition, error) {
	client := ecs.NewFromConfig(awsConfig)

	out1, err := client.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
		Services: []string{serviceName},
		Cluster:  aws.String(clusterArn),
	})
	if err != nil {
		return nil, err
	}
	if len(out1.Services) < 1 {
		return nil, fmt.Errorf("could not find service %q in cluster %q", serviceName, clusterArn)
	}

	out2, err := client.DescribeTaskDefinition(context.Background(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: out1.Services[0].TaskDefinition,
	})
	if err != nil {
		return nil, err
	}
	return out2.TaskDefinition, nil
}
