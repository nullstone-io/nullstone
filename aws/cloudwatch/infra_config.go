package cloudwatch

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/fatih/color"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws_cloudwatch"
	"io"
	"log"
	"time"
)

var (
	DefaultWatchInterval = 1 * time.Second
	bold = color.New(color.Bold)
	normal = color.New()
)

type MessageEmitter func(event cwltypes.FilteredLogEvent)

type InfraConfig struct {
	Outputs aws_cloudwatch.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger.Printf("region: %q\n", c.Outputs.Region)
	logger.Printf("log group: %q\n", c.Outputs.LogGroupName)
}

func (c InfraConfig) StreamLogs(ctx context.Context, options config.LogStreamOptions, w io.Writer) error {
	emitter := func(event cwltypes.FilteredLogEvent) {
		timestamp := time.Unix (*event.Timestamp / 1000, 0)
		normal.Fprintf(w, "%s ", timestamp.Format("RFC822"))
		bold.Fprintf(w, "[%s]", *event.LogStreamName)
		normal.Fprintf(w, " %s", *event.Message)
		normal.Println()
	}
	fn := c.writeLatestEvents(options, emitter)

	if options.WatchInterval == time.Duration(0) {
		options.WatchInterval = DefaultWatchInterval
	}

	for {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("error querying logs: %w", err)
		}
		if options.WatchInterval < 0 {
			// A negative watch interval indicates
			return nil
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(options.WatchInterval):
		}
	}
}

// Each pass of writeLatestEvents will emit all events (based on filtering)
// We record the last event timestamp every time we emit an event
// This allows us to pick up where we left off from a previous query
func (c InfraConfig) writeLatestEvents(options config.LogStreamOptions, emitter MessageEmitter) func(ctx context.Context) error {
	cwlClient := cloudwatchlogs.NewFromConfig(nsaws.NewConfig(c.Outputs.LogReader, c.Outputs.Region))
	input := cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:  aws.String(c.Outputs.LogGroupName),
		NextToken:     nil,
		StartTime:     toAwsTime(options.StartTime),
		EndTime:       toAwsTime(options.EndTime),
		FilterPattern: options.Pattern,
	}
	var lastEventTime *int64

	return func(ctx context.Context) error {
		if lastEventTime == nil {
			input.StartTime = lastEventTime
		}
		input.NextToken = nil
		for {
			out, err := cwlClient.FilterLogEvents(ctx, &input)
			if err != nil {
				return fmt.Errorf("error filtering log events: %w", err)
			}
			for _, event := range out.Events {
				lastEventTime = event.Timestamp
				emitter(event)
			}
			input.NextToken = out.NextToken
			if out.NextToken == nil {
				break
			}
		}
		return nil
	}
}

func toAwsTime(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	return aws.Int64(t.Unix() * int64(time.Millisecond))
}
