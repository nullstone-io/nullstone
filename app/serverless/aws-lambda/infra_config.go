package aws_lambda

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
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
	logger = log.New(logger.Writer(), "    ", 0)
	logger.Printf("lambda arn: %q\n", c.Outputs.LambdaArn)
	logger.Printf("lambda name: %q\n", c.Outputs.LambdaName)
	logger.Printf("artifacts bucket: %q\n", c.Outputs.ArtifactsBucketName)
}

func (c InfraConfig) UploadArtifact(ctx context.Context, content io.ReadSeeker, version string) error {
	s3Client := s3.NewFromConfig(nsaws.NewConfig(c.Outputs.Deployer, c.Outputs.Region))

	// Calculate md5 content to add as header (necessary for s3 buckets that have object lock enabled)
	// After calculating, we need to reset the content stream to transmit using s3.PutObject
	md5Summer := md5.New()
	if _, err := io.Copy(md5Summer, content); err != nil {
		return fmt.Errorf("error calculating md5 hash: %w", err)
	}
	md5Sum := hex.EncodeToString(md5Summer.Sum(nil))
	if _, err := content.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("error resetting uploaded content after calculating md5 hash: %w", err)
	}

	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:     aws.String(c.Outputs.ArtifactsBucketName),
		Key:        aws.String(c.Outputs.ArtifactsKey(version)),
		Body:       content,
		ContentMD5: aws.String(md5Sum),
	})
	return err
}

func (c InfraConfig) UpdateLambdaVersion(ctx context.Context, version string) error {
	λClient := lambda.NewFromConfig(nsaws.NewConfig(c.Outputs.Deployer, c.Outputs.Region))
	_, err := λClient.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(c.Outputs.LambdaName),
		DryRun:       false,
		Publish:      false,
		S3Bucket:     aws.String(c.Outputs.ArtifactsBucketName),
		S3Key:        aws.String(c.Outputs.ArtifactsKey(version)),
	})
	return err
}
