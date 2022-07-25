package ec2

import (
	"context"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ssm"
)

func ExecCommand(ctx context.Context, infra Outputs, cmd string, parameters map[string][]string) error {
	// TODO: Add support for cmd
	region := infra.Region
	awsConfig := nsaws.NewConfig(infra.Adminer, region)
	return ssm.StartEc2Session(ctx, awsConfig, region, infra.InstanceId, parameters)
}
