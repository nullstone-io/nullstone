package nsaws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/smithy-go/logging"
	caws "gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
	"os"
)

const (
	DefaultAwsRegion = "us-east-1"
	AwsTraceEnvVar   = "AWS_TRACE"
)

func NewConfig(user caws.User, region string) aws.Config {
	awsConfig := aws.Config{}
	if os.Getenv(AwsTraceEnvVar) != "" {
		awsConfig.Logger = logging.NewStandardLogger(os.Stderr)
		awsConfig.ClientLogMode = aws.LogRequestWithBody | aws.LogResponseWithBody
	}
	awsConfig.Region = DefaultAwsRegion
	if region != "" {
		awsConfig.Region = region
	}
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider(user.AccessKeyId, user.SecretAccessKey, "")
	return awsConfig
}
