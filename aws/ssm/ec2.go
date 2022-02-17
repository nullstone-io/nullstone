package ssm

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// StartEc2Session initiates an interactive SSH session with an EC2 instance using SSM
// See setup guide: https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-getting-started.html
// In short, the following is necessary for this function to work
//   - ec2 instance has SSM agent installed and registered
//   - ec2 instance is configured with instance profile that has the AmazonSSMManagedInstanceCore policy attached (or an equivalent custom policy)
//   - config contains an AWS identity that has access to ssm:StartSession on the EC2 Instance
func StartEc2Session(ctx context.Context, config aws.Config, region, instanceId string) error {
	ssmClient := ssm.NewFromConfig(config)
	input := &ssm.StartSessionInput{
		Target: aws.String(instanceId),
		Reason: aws.String("nullstone exec"),
	}
	out, err := ssmClient.StartSession(ctx, input)
	if err != nil {
		return fmt.Errorf("error starting ssm session: %w", err)
	}

	target := ssm.StartSessionInput{
		Target: aws.String(instanceId),
	}

	er := ec2.NewDefaultEndpointResolver()
	endpoint, _ := er.ResolveEndpoint(region, ec2.EndpointResolverOptions{})

	return StartSession(ctx, out, target, region, endpoint.URL)
}
