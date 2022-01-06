package aws_fargate

import (
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
)

type Outputs struct {
	ClusterArn string `ns:"cluster_arn"`

	// Deprecated: Deployer was moved to the fargate service in module v0.11.0+
	// Remove "optional" from deployer in fargate service when removing
	Deployer aws.User `ns:"deployer,optional"`
}
