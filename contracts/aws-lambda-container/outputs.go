package aws_lambda_container

import (
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
)

type Outputs struct {
	Region       string          `ns:"region"`
	Deployer     aws.User        `ns:"deployer"`
	LambdaArn    string          `ns:"lambda_arn"`
	LambdaName   string          `ns:"lambda_name"`
	ImageRepoUrl docker.ImageUrl `ns:"image_repo_url,optional"`
	ImagePusher  aws.User        `ns:"image_pusher,optional"`
}
