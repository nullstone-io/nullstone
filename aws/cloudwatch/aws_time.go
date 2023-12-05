package cloudwatch

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"strconv"
	"time"
)

func toAwsTime(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	return aws.Int64(t.UnixNano() / int64(time.Millisecond))
}

func fromAwsTimeString(s string) time.Time {
	ms, _ := strconv.ParseInt(s, 10, 64)
	return time.Unix(ms/1000, 0)
}
