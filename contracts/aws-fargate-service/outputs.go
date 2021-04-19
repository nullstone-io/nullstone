package aws_fargate_service

import (
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
	aws_fargate "gopkg.in/nullstone-io/nullstone.v0/contracts/aws-fargate"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
)

type Outputs struct {
	ServiceName  string          `ns:"service_name"`
	ImageRepoUrl docker.ImageUrl `ns:"image_repo_url,optional"`
	ImagePusher  aws.User        `ns:"image_pusher,optional"`

	Cluster aws_fargate.Outputs `ns:",connectionType:cluster/aws-fargate"`
}
