package cloudwatch

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"sync"
	"time"
)

func queryLogs(infra Outputs, options config.LogStreamOptions, emitter MessageEmitter) func(ctx context.Context) error {
	query := "@log, @logStream, @message, @timestamp | sort @timestamp asc"
	cwlClient := cloudwatchlogs.NewFromConfig(nsaws.NewConfig(infra.LogReader, infra.Region))
	now := time.Now()
	if options.StartTime == nil {
		options.StartTime = &now
	}
	if options.EndTime == nil {
		options.EndTime = &now
	}
	input := &cloudwatchlogs.StartQueryInput{
		QueryString: aws.String(query),
		StartTime:   toAwsTime(options.StartTime),
		EndTime:     toAwsTime(options.EndTime),
	}

	var lastEventTime *time.Time
	loadLoadGroupNames := sync.OnceValues(func() ([]string, error) {
		return ExpandLogGroups(context.Background(), infra)
	})

	return func(ctx context.Context) error {
		logGroupNames, err := loadLoadGroupNames()
		if err != nil {
			return err
		}
		input.LogGroupNames = logGroupNames
		if lastEventTime != nil {
			input.StartTime = toAwsTime(lastEventTime)
		}

		out, err := cwlClient.StartQuery(ctx, input)
		if err != nil {
			return fmt.Errorf("error querying log events: %w", err)
		}

		results, err := cwlClient.GetQueryResults(ctx, &cloudwatchlogs.GetQueryResultsInput{QueryId: out.QueryId})
		if err != nil {
			return fmt.Errorf("error retrieving log query results: %w", err)
		}
		if results.Status != types.QueryStatusComplete {
			return nil
		}
		for _, result := range results.Results {
			le := LogEventFromQueryResult(result)
			lastEventTime = &le.Timestamp
			emitter(le)
		}
		return nil
	}
}
