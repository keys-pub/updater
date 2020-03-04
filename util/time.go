package util

import "time"

// TimeMs is time as number of milliseconds from epoch.
type TimeMs int64

// TimeToMillis returns milliseconds since epoch from time.Time.
// If t.IsZero() we return 0.
func TimeToMillis(t time.Time) TimeMs {
	if t.IsZero() {
		return 0
	}
	return TimeMs(t.UnixNano() / int64(time.Millisecond))
}

// TimePtrToMillis returns milliseconds since epoch from time.Time.
// If t is nil or t.IsZero() we return 0.
func TimePtrToMillis(t *time.Time) TimeMs {
	if t == nil {
		return 0
	}
	return TimeToMillis(*t)
}

// TimeFromMillis returns time.Time from milliseconds since epoch.
func TimeFromMillis(m TimeMs) time.Time {
	if m == 0 {
		return time.Time{}
	}
	return time.Unix(0, int64(m)*int64(time.Millisecond))
}
