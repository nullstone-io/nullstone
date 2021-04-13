package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type DeployerUser struct {
	Name            string `json:"name"`
	AccessKeyId     string `json:"access_key"`
	SecretAccessKey string `json:"secret_key"`
}

func (u DeployerUser) CreateConfig() aws.Config {
	awsConfig := aws.Config{}
	// TODO: How do we set the region?
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider(u.AccessKeyId, u.SecretAccessKey, "")
	return awsConfig
}
