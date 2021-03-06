package aws_fargate_service

import (
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
	aws_fargate "gopkg.in/nullstone-io/nullstone.v0/contracts/aws-fargate"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
)

type Outputs struct {
	Region            string          `ns:"region"`
	ServiceName       string          `ns:"service_name"`
	ImageRepoUrl      docker.ImageUrl `ns:"image_repo_url,optional"`
	ImagePusher       aws.User        `ns:"image_pusher,optional"`
	MainContainerName string          `ns:"main_container_name,optional"`
	Deployer          aws.User        `ns:"deployer,optional"`

	Cluster aws_fargate.Outputs `ns:",connectionType:cluster/aws-fargate,connectionContract=cluster/aws/ecs:fargate"`
}

func (o Outputs) GetDeployer() aws.User {
	if o.Deployer.Name == "" {
		return o.Cluster.Deployer
	}
	return o.Deployer
}
