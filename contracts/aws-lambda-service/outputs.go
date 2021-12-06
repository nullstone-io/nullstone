package aws_lambda_service

import (
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"strings"
)

const (
	KeyTemplateAppVersion = "{{app-version}}"
)

type Outputs struct {
	Region     string   `ns:"region"`
	Deployer   aws.User `ns:"deployer"`
	LambdaArn  string   `ns:"lambda_arn"`
	LambdaName string   `ns:"lambda_name"`

	ArtifactSource       string          `ns:"artifact_source"`
	ArtifactsBucketName  string          `ns:"artifacts_bucket_name,optional"`
	ArtifactsKeyTemplate string          `ns:"artifacts_key_template,optional"`
	ImageRepoUrl         docker.ImageUrl `ns:"image_repo_url,optional"`
	ImagePusher          aws.User        `ns:"image_pusher,optional"`
}

func (o Outputs) ArtifactsKey(appVersion string) string {
	return strings.Replace(o.ArtifactsKeyTemplate, KeyTemplateAppVersion, appVersion, -1)
}
