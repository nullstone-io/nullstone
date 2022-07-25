package ecs

import (
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
)

type Outputs struct {
	Region            string     `ns:"region"`
	ServiceName       string     `ns:"service_name"`
	TaskArn           string     `ns:"task_arn"`
	MainContainerName string     `ns:"main_container_name,optional"`
	Deployer          nsaws.User `ns:"deployer,optional"`

	Cluster ClusterOutputs `ns:",connectionType:cluster/aws-ecs,connectionContract:cluster/aws/ecs:*"`
}

type ClusterOutputs struct {
	ClusterArn string `ns:"cluster_arn"`
}
