package cloudwatch

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"strings"
	"time"
)

type LogEvent struct {
	LogGroupQualifiedName string // <account-id>:<log-group-name>
	LogGroupName          string // <log-group-name>
	LogStreamName         string
	Message               string
	RawTimestamp          string
	Timestamp             time.Time
}

func LogEventFromFilteredLogEvent(event cwltypes.FilteredLogEvent) LogEvent {
	return LogEvent{
		LogStreamName: aws.ToString(event.LogStreamName),
		Message:       aws.ToString(event.Message),
		Timestamp:     time.Unix(*event.Timestamp/1000, 0),
	}
}

func LogEventFromQueryResult(result []cwltypes.ResultField) LogEvent {
	le := LogEvent{}
	for _, kvp := range result {
		val := aws.ToString(kvp.Value)
		switch aws.ToString(kvp.Field) {
		case "@log":
			le.LogGroupQualifiedName = val
		case "@logStream":
			le.LogStreamName = val
		case "@message":
			le.Message = val
		case "@timestamp":
			le.RawTimestamp = val
		}
	}
	_, le.LogGroupName, _ = strings.Cut(le.LogGroupQualifiedName, ":") // <log-group-name>
	le.Timestamp = fromAwsTimeString(le.RawTimestamp)
	return le
}
