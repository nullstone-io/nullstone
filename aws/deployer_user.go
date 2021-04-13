package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const (
	DefaultAwsRegion = "us-east-1"
)

// DeployerUser contains credentials for a user that has access to deploy a particular app
// This structure must match the fields defined in outputs of the module
type DeployerUser struct {
	Name            string `json:"name"`
	AccessKeyId     string `json:"access_key"`
	SecretAccessKey string `json:"secret_key"`
}

func (u DeployerUser) CreateConfig() aws.Config {
	awsConfig := aws.Config{}
	awsConfig.Region = DefaultAwsRegion
	// TODO: How do we set the region?
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider(u.AccessKeyId, u.SecretAccessKey, "")
	return awsConfig
}
