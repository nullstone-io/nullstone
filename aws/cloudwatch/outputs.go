package cloudwatch

import (
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
)

type Outputs struct {
	Region       string     `ns:"region"`
	LogReader    nsaws.User `ns:"log_reader"`
	LogGroupName string     `ns:"log_group_name"`
}
