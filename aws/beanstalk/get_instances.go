package beanstalk

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func GetInstances(ctx context.Context, infra Outputs) ([]string, error) {
	ec2Client := ec2.NewFromConfig(infra.AdminerConfig())
	// TODO: Get Instance
	//out, err := ec2Client.ListTasks(ctx, &ecs.ListTasksInput{
	//	Cluster:     aws.String(infra.ClusterArn()),
	//	ServiceName: aws.String(infra.ServiceName),
	//})
	if err != nil {
		return nil, err
	}
	return out.InstanceIds, nil
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
