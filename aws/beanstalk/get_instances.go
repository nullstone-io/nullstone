package beanstalk

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
)

func GetInstances(ctx context.Context, infra Outputs) ([]string, error) {
	client := elasticbeanstalk.NewFromConfig(infra.AdminerConfig())
	out, err := client.DescribeEnvironmentResources(ctx, &elasticbeanstalk.DescribeEnvironmentResourcesInput{
		EnvironmentId: aws.String(infra.EnvironmentId),
	})
	if err != nil {
		return nil, err
	}
	instanceIds := make([]string, 0)
	for _, instance := range out.EnvironmentResources.Instances {
		instanceIds = append(instanceIds, *instance.Id)
	}
	return instanceIds, nil
}

func GetRandomInstance(ctx context.Context, infra Outputs) (string, error) {
	instanceIds, err := GetInstances(ctx, infra)
	if err != nil {
		return "", err
	}

	for _, instanceId := range instanceIds {
		return instanceId, nil
	}
	return "", nil
}
