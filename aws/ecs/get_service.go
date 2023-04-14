package ecs

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
)

func GetService(ctx context.Context, infra Outputs) (*ecstypes.Service, error) {
	ecsClient := ecs.NewFromConfig(nsaws.NewConfig(infra.Deployer, infra.Region))
	out, err := ecsClient.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Services: []string{infra.ServiceName},
		Cluster:  aws.String(infra.ClusterArn()),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Services) > 0 {
		return &out.Services[0], nil
	}
	return nil, nil
}
