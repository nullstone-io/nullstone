package ec2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
)

type Outputs struct {
	Region     string     `ns:"region"`
	InstanceId string     `ns:"instance_id"`
	Adminer    nsaws.User `ns:"adminer,optional"`
}

func (o Outputs) AdminerConfig() aws.Config {
	return nsaws.NewConfig(o.Adminer, o.Region)
}
