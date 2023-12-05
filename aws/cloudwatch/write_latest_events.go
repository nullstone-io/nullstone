package cloudwatch

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

// Each pass of writeLatestEvents will emit all events (based on filtering)
// We record the last event timestamp every time we emit an event
// This allows us to pick up where we left off from a previous query
func writeLatestEvents(infra Outputs, logGroupName string, options config.LogStreamOptions, emitter MessageEmitter) func(ctx context.Context) error {
	cwlClient := cloudwatchlogs.NewFromConfig(nsaws.NewConfig(infra.LogReader, infra.Region))
	input := cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		NextToken:     nil,
		StartTime:     toAwsTime(options.StartTime),
		EndTime:       toAwsTime(options.EndTime),
		FilterPattern: options.Pattern,
	}
	var lastEventTime *int64
	visitedEventIds := map[string]bool{}

	return func(ctx context.Context) error {
		if lastEventTime != nil {
			input.StartTime = lastEventTime
		}
		input.NextToken = nil
		for {
			out, err := cwlClient.FilterLogEvents(ctx, &input)
			if err != nil {
				return fmt.Errorf("error filtering log events: %w", err)
			}
			for _, event := range out.Events {
				if _, ok := visitedEventIds[*event.EventId]; ok {
					continue
				}
				lastEventTime = event.Timestamp
				visitedEventIds[*event.EventId] = true
				emitter(LogEventFromFilteredLogEvent(event))
			}
			input.NextToken = out.NextToken
			if out.NextToken == nil {
				break
			}
		}
		return nil
	}
}
