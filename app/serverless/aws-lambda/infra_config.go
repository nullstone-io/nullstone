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

// InfraConfig provides a minimal understanding of the infrastructure provisioned for a module type=aws-fargate
type InfraConfig struct {
	Outputs aws_lambda_service.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger.Printf("lambda arn: %q\n", c.Outputs.LambdaArn)
	logger.Printf("lambda name: %q\n", c.Outputs.LambdaName)
	logger.Printf("artifacts bucket: %q\n", c.Outputs.ArtifactsBucketName)
}

func (c InfraConfig) UploadArtifact(ctx context.Context, content io.Reader, version string) error {
	s3Client := s3.NewFromConfig(nsaws.NewConfig(c.Outputs.Deployer))
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.Outputs.ArtifactsBucketName),
		Key:    aws.String(c.Outputs.ArtifactsKey(version)),
		Body:   content,
	})
	return err
}

func (c InfraConfig) UpdateLambdaVersion(ctx context.Context, version string) error {
	λClient := lambda.NewFromConfig(nsaws.NewConfig(c.Outputs.Deployer))
	_, err := λClient.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(c.Outputs.LambdaName),
		DryRun:       false,
		Publish:      false,
		S3Bucket:     aws.String(c.Outputs.ArtifactsBucketName),
		S3Key:        aws.String(c.Outputs.ArtifactsKey(version)),
	})
	return err
}
