package aws_lambda_container

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"gopkg.in/nullstone-io/nullstone.v0/aws"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws-lambda-container"
	"log"
)

type InfraConfig struct {
	Outputs aws_lambda_container.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger = log.New(logger.Writer(), "    ", 0)
	logger.Printf("lambda arn: %q\n", c.Outputs.LambdaArn)
	logger.Printf("lambda name: %q\n", c.Outputs.LambdaName)
	logger.Printf("repository image url: %q\n", c.Outputs.ImageRepoUrl)
}

func (c InfraConfig) UpdateLambdaVersion(ctx context.Context, version string) error {
	λClient := lambda.NewFromConfig(nsaws.NewConfig(c.Outputs.Deployer, c.Outputs.Region))
	imageUrl := c.Outputs.ImageRepoUrl
	imageUrl.Digest = ""
	imageUrl.Tag = version
	_, err := λClient.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(c.Outputs.LambdaName),
		DryRun:       false,
		Publish:      false,
		ImageUri:     aws.String(imageUrl.String()),
	})
	return err
}
