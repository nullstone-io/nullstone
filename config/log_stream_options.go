package config

import (
	"io"
	"time"
)

type LogStreamOptions struct {
	StartTime *time.Time
	EndTime   *time.Time
	// The filter pattern to use. For more information, see Filter and Pattern Syntax
	// (https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html).
	// If not provided, all the events are matched.
	Pattern *string

	// WatchInterval dictates how often the log streamer will query AWS for new events
	// If left unspecified or 0, will use default watch interval of 1s
	// If a negative value is specified, watching will disable, the log streamer will terminate as soon as logs are emitted
	WatchInterval time.Duration

	// Out defines a colorized output stream to stream logs
	Out io.Writer
}
