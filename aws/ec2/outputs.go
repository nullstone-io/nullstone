package ec2

import (
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
)

type Outputs struct {
	Region     string     `ns:"region"`
	InstanceId string     `ns:"instance_id"`
	Adminer    nsaws.User `ns:"adminer,optional"`
}
