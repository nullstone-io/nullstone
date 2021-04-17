package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/smithy-go/logging"
	"os"
)

const (
	DefaultAwsRegion = "us-east-1"
	AwsTraceEnvVar   = "AWS_TRACE"
)

// ActionUser contains credentials for a user that has access to perform a particular action in AWS
// This structure must match the fields defined in outputs of the module
type ActionUser struct {
	Name            string `json:"name"`
	AccessKeyId     string `json:"access_key"`
	SecretAccessKey string `json:"secret_key"`
}

func (u ActionUser) CreateConfig() aws.Config {
	awsConfig := aws.Config{}
	if os.Getenv(AwsTraceEnvVar) != "" {
		awsConfig.Logger = logging.NewStandardLogger(os.Stderr)
		awsConfig.ClientLogMode = aws.LogRequestWithBody | aws.LogResponseWithBody
	}
	awsConfig.Region = DefaultAwsRegion
	// TODO: How do we set the region?
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider(u.AccessKeyId, u.SecretAccessKey, "")
	return awsConfig
}
