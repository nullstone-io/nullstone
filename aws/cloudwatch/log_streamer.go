package cloudwatch

import (
	"context"
	"fmt"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/fatih/color"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"gopkg.in/nullstone-io/nullstone.v0/display"
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

func NewLogStreamer(osWriters logging.OsWriters, nsConfig api.Config, appDetails app.Details) (admin.LogStreamer, error) {
	outs, err := outputs.Retrieve[Outputs](nsConfig, appDetails.Workspace)
	if err != nil {
		return nil, err
	}

	return LogStreamer{
		OsWriters: osWriters,
		Details:   appDetails,
		Infra:     outs,
	}, nil
}

type LogStreamer struct {
	OsWriters logging.OsWriters
	Details   app.Details
	Infra     Outputs
}

func (l LogStreamer) Stream(ctx context.Context, options config.LogStreamOptions) error {
	stdout := l.OsWriters.Stdout()

	emitter := func(event cwltypes.FilteredLogEvent) {
		timestamp := time.Unix(*event.Timestamp/1000, 0)
		normal.Fprintf(stdout, "%s ", display.FormatTime(timestamp))
		bold.Fprintf(stdout, "[%s]", *event.LogStreamName)
		normal.Fprintf(stdout, " %s", *event.Message)
		normal.Fprintln(stdout)
	}
	fn := writeLatestEvents(l.Infra, options, emitter)

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
