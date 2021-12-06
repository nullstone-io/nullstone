package aws_lambda

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	aws_lambda_service "gopkg.in/nullstone-io/nullstone.v0/contracts/aws-lambda-service"
	"io"
	"log"
)

const (
	ArtifactSourceS3     = "s3"
	ArtifactSourceDocker = "docker"
)

// InfraConfig provides a minimal understanding of the infrastructure provisioned for a module type=aws-fargate
type InfraConfig struct {
	Outputs aws_lambda_service.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger = log.New(logger.Writer(), "    ", 0)
	logger.Printf("lambda arn: %q\n", c.Outputs.LambdaArn)
	logger.Printf("lambda name: %q\n", c.Outputs.LambdaName)
	logger.Printf("artifact source: %q\n", c.Outputs.ArtifactSource)
	if c.HasDockerArtifactSource() {
		logger.Printf("repository image url: %q\n", c.Outputs.ImageRepoUrl)
	} else {
		logger.Printf("artifacts bucket: %q\n", c.Outputs.ArtifactsBucketName)
	}
}

func (c InfraConfig) UploadArtifact(ctx context.Context, content io.Reader, version string) error {
	s3Client := s3.NewFromConfig(nsaws.NewConfig(c.Outputs.Deployer, c.Outputs.Region))
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.Outputs.ArtifactsBucketName),
		Key:    aws.String(c.Outputs.ArtifactsKey(version)),
		Body:   content,
	})
	return err
}

func (c InfraConfig) UpdateLambdaVersion(ctx context.Context, version string) error {
	λClient := lambda.NewFromConfig(nsaws.NewConfig(c.Outputs.Deployer, c.Outputs.Region))
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(c.Outputs.LambdaName),
		DryRun:       false,
		Publish:      false,
	}
	if c.HasDockerArtifactSource() {
		imageUrl := c.Outputs.ImageRepoUrl
		imageUrl.Tag = version
		input.ImageUri = aws.String(imageUrl.String())
	} else {
		input.S3Bucket = aws.String(c.Outputs.ArtifactsBucketName)
		input.S3Key = aws.String(c.Outputs.ArtifactsKey(version))
	}
	_, err := λClient.UpdateFunctionCode(ctx, input)
	return err
}

func (c InfraConfig) HasDockerArtifactSource() bool {
	return c.Outputs.ArtifactSource == ArtifactSourceDocker
}
