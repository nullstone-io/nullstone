package cloudwatch

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"strings"
)

func ExpandLogGroups(ctx context.Context, infra Outputs) ([]string, error) {
	before, found := strings.CutSuffix(infra.LogGroupName, "*")
	if !found {
		return []string{infra.LogGroupName}, nil
	}

	cwlClient := cloudwatchlogs.NewFromConfig(nsaws.NewConfig(infra.LogReader, infra.Region))
	out, err := cwlClient.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(before),
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving a listing of log groups: %w", err)
	}
	results := make([]string, 0)
	for _, lg := range out.LogGroups {
		results = append(results, aws.ToString(lg.LogGroupName))
	}
	return results, nil
}
