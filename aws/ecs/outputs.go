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

	Cluster          ClusterOutputs          `ns:",connectionContract:cluster/aws/ecs:*,optional"`
	ClusterNamespace ClusterNamespaceOutputs `ns:",connectionContract:cluster-namespace/aws/ecs:*,optional"`
}

func (o Outputs) ClusterArn() string {
	if o.ClusterNamespace.ClusterArn != "" {
		return o.ClusterNamespace.ClusterArn
	}
	return o.Cluster.ClusterArn
}

type ClusterNamespaceOutputs struct {
	ClusterArn string `ns:"cluster_arn"`
}

type ClusterOutputs struct {
	ClusterArn string `ns:"cluster_arn"`
}
