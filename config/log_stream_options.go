package config

import (
	"fmt"
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

func (o LogStreamOptions) QueryTimeMessage() string {
	if o.StartTime != nil {
		if o.EndTime != nil {
			return fmt.Sprintf("Querying logs between %s and %s", o.StartTime.Format(time.RFC822), o.EndTime.Format(time.RFC822))
		}
		return fmt.Sprintf("Querying logs starting %s", o.StartTime.Format(time.RFC822))
	} else if o.EndTime != nil {
		return fmt.Sprintf("Querying logs until %s", o.EndTime.Format(time.RFC822))
	}
	return fmt.Sprintf("Querying all logs")
}

func (o LogStreamOptions) WatchMessage() string {
	wi := o.WatchInterval
	if wi < 0 {
		return "Not watching logs"
	}
	if wi == 0 {
		wi = time.Second
	}
	return fmt.Sprintf("Watching logs (poll interval = %s)", wi)
}
