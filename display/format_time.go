package display

import (
	"time"
)

const (
	CommonFormat = "Mon Jan _2 15:04:05 MST"
)

func FormatTime(t time.Time) string {
	return t.In(time.Local).Format(CommonFormat)
}

func FormatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.In(time.Local).Format(CommonFormat)
}
