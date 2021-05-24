package aws_ecr

import (
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
)

type Outputs struct {
	Region       string          `ns:"region,optional"`
	ImageRepoUrl docker.ImageUrl `ns:"image_repo_url,optional"`
	ImagePusher  aws.User        `ns:"image_pusher,optional"`
}
