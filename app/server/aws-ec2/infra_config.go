package aws_ec2

import (
	"context"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ssm"
	aws_ec2 "gopkg.in/nullstone-io/nullstone.v0/contracts/aws-ec2"
	"log"
)

type InfraConfig struct {
	Outputs aws_ec2.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger = log.New(logger.Writer(), "    ", 0)
	logger.Printf("instance id: %q\n", c.Outputs.InstanceId)
}

func (c InfraConfig) ExecCommand(ctx context.Context, cmd string) error {
	// TODO: Add support for cmd
	region := c.Outputs.Region
	awsConfig := nsaws.NewConfig(c.Outputs.Adminer, region)
	return ssm.StartEc2Session(ctx, awsConfig, region, c.Outputs.InstanceId)
}
