package aws_lambda_service

import (
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
	"strings"
)

const (
	KeyTemplateAppVersion = "{{app-version}}"
)

type Outputs struct {
	Region               string   `ns:"region"`
	Deployer             aws.User `ns:"deployer"`
	LambdaArn            string   `ns:"lambda_arn"`
	LambdaName           string   `ns:"lambda_name"`
	ArtifactsBucketName  string   `ns:"artifacts_bucket_name"`
	ArtifactsKeyTemplate string   `ns:"artifacts_key_template"`
}

func (o Outputs) ArtifactsKey(appVersion string) string {
	return strings.Replace(o.ArtifactsKeyTemplate, KeyTemplateAppVersion, appVersion, -1)
}
