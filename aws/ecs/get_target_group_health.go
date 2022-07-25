package ecs

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
)

func GetTargetGroupHealth(ctx context.Context, infra Outputs, targetGroupArn string) ([]elbv2types.TargetHealthDescription, error) {
	elbClient := elasticloadbalancingv2.NewFromConfig(nsaws.NewConfig(infra.Deployer, infra.Region))
	out, err := elbClient.DescribeTargetHealth(ctx, &elasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(targetGroupArn),
	})
	if err != nil {
		return nil, err
	}
	return out.TargetHealthDescriptions, nil
}
