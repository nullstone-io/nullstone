package aws_cloudwatch

import "gopkg.in/nullstone-io/nullstone.v0/contracts/aws"

type Outputs struct {
	Region       string   `ns:"region"`
	LogReader    aws.User `ns:"log_reader"`
	LogGroupName string   `ns:"log_group_name"`
}
