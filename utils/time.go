package utils

import (
	"fmt"
	"time"
)

// Create a relative time string from now to past (e.g., "1h ago")
func RelativeTimeStr(t time.Time) string {
	return RelativeTimeStrFrom(t, time.Now())
}

// Create a relative time string from `asOf` back to `t` (e.g., "1h ago").
// Use this when the reference point must be pinned (e.g. a snapshot taken at job
// start) so that "X ago" stays consistent with the data being displayed.
func RelativeTimeStrFrom(t time.Time, asOf time.Time) string {
	diff := asOf.Sub(t)
	diffInSeconds := int(diff.Seconds())
	if diffInSeconds < 2 {
		return "Just now"
	}
	if diffInSeconds < 60 {
		return fmt.Sprint(diffInSeconds, "s ago")
	}
	if diffInSeconds < 60*60 {
		return fmt.Sprint(diffInSeconds/60, "m ago")
	}
	if diffInSeconds < 60*60*24 {
		return fmt.Sprint(diffInSeconds/(60*60), "h ago")
	}
	return fmt.Sprint(diffInSeconds/(60*60*24), "d ago")
}

// Parse an iso string into a time.Time
// e.g. 2023-01-27T14:40:53.858Z
// represents javascript toISOString()
func ParseIsoTime(isoTime string) (time.Time, error) {
	return time.Parse(time.RFC3339, isoTime)
}

func TimeToIsoString(ts time.Time) string {
	return ts.Format(time.RFC3339Nano)
}

func SecondsSinceEpochToTime(seconds int64) time.Time {
	return time.Unix(seconds, 0)
}
