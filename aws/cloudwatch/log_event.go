package cloudwatch

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"time"
)

type LogEvent struct {
	LogStreamName string
	Message       string
	Timestamp     time.Time
}

func LogEventFromFilteredLogEvent(event cwltypes.FilteredLogEvent) LogEvent {
	return LogEvent{
		LogStreamName: aws.ToString(event.LogStreamName),
		Message:       aws.ToString(event.Message),
		Timestamp:     time.Unix(*event.Timestamp/1000, 0),
	}
}
