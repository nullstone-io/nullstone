package ssm

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func StartEc2Session(config aws.Config, region, instanceId string) error {
	ctx := context.Background()

	ssmClient := ssm.NewFromConfig(config)
	input := &ssm.StartSessionInput{
		Target:       aws.String(instanceId),
		DocumentName: aws.String("AWS-StartSSHSession"),
		Reason:       aws.String("nullstone exec"),
	}
	out, err := ssmClient.StartSession(ctx, input)
	if err != nil {
		return fmt.Errorf("error establishing ecs execute command: %w", err)
	}

	er := ec2.NewDefaultEndpointResolver()
	endpoint, _ := er.ResolveEndpoint(region, ec2.EndpointResolverOptions{})

	return StartSession(ctx, out, region, instanceId, endpoint.URL)
}
