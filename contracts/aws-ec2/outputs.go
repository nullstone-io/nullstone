package aws_ec2

import "gopkg.in/nullstone-io/nullstone.v0/contracts/aws"

type Outputs struct {
	Region     string   `ns:"region"`
	InstanceId string   `ns:"instance_id"`
	Adminer    aws.User `ns:"adminer,optional"`
}
