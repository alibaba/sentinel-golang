package util

import (
	"time"
)

const (
	DateFormat         = "2006-01-02 15:04:05"
	UnixTimeUnitOffset = uint64(time.Millisecond / time.Nanosecond)
)

func FormatTimeMillis(ts uint64) string {
	return time.Unix(0, int64(ts*UnixTimeUnitOffset)).Format(DateFormat)
}

// Returns the current Unix timestamp in milliseconds.
func CurrentTimeMillis() uint64 {
	return uint64(time.Now().UnixNano()) / UnixTimeUnitOffset
}
