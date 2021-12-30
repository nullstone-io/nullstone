package cloudwatch

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/fatih/color"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/app_logs"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws-cloudwatch"
	"gopkg.in/nullstone-io/nullstone.v0/nsaws"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
	"os"
	"time"
)

var (
	logger               = log.New(os.Stderr, "", 0)
	infoLogger           = log.New(logger.Writer(), "    ", 0)
	DefaultWatchInterval = 1 * time.Second
	bold                 = color.New(color.Bold)
	normal               = color.New()
)

type MessageEmitter func(event cwltypes.FilteredLogEvent)

var _ app_logs.Provider = Provider{}

type Provider struct {
}

func (p Provider) identify(nsConfig api.Config, details app.Details) (*aws_cloudwatch.Outputs, error) {
	logger.Printf("Retrieving log provider details for app %q\n", details.App.Name)
	retriever := outputs.Retriever{NsConfig: nsConfig}
	var cwOutputs aws_cloudwatch.Outputs
	if err := retriever.Retrieve(details.Workspace, &cwOutputs); err != nil {
		return nil, fmt.Errorf("Unable to retrieve app logger details: %w", err)
	}
	infoLogger.Printf("region: %q\n", cwOutputs.Region)
	infoLogger.Printf("log group: %q\n", cwOutputs.LogGroupName)
	return &cwOutputs, nil
}

func (p Provider) Stream(ctx context.Context, nsConfig api.Config, details app.Details, options config.LogStreamOptions) error {
	cwOutputs, err := p.identify(nsConfig, details)
	if err != nil {
		return err
	}

	emitter := func(event cwltypes.FilteredLogEvent) {
		timestamp := time.Unix(*event.Timestamp/1000, 0)
		normal.Fprintf(options.Out, "%s ", timestamp.Format(time.RFC822Z))
		bold.Fprintf(options.Out, "[%s]", *event.LogStreamName)
		normal.Fprintf(options.Out, " %s", *event.Message)
		normal.Fprintln(options.Out)
	}
	fn := p.writeLatestEvents(*cwOutputs, options, emitter)

	if options.WatchInterval == time.Duration(0) {
		options.WatchInterval = DefaultWatchInterval
	}

	logger.Println(options.QueryTimeMessage())
	logger.Println(options.WatchMessage())
	logger.Println()
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
func (p Provider) writeLatestEvents(cwOutputs aws_cloudwatch.Outputs, options config.LogStreamOptions, emitter MessageEmitter) func(ctx context.Context) error {
	cwlClient := cloudwatchlogs.NewFromConfig(nsaws.NewConfig(cwOutputs.LogReader, cwOutputs.Region))
	input := cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:  aws.String(cwOutputs.LogGroupName),
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
	return aws.Int64(t.UnixNano() / int64(time.Millisecond))
}
