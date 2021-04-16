package aws_fargate

import (
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
)

type Outputs struct {
	ClusterArn string   `ns:"cluster_arn"`
	Deployer   aws.User `ns:"deployer"`
}
