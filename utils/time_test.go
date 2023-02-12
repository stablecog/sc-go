package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRelativeTimeStr(t *testing.T) {
	// Now
	now := time.Now()
	assert.Equal(t, "Just now", RelativeTimeStr(now))
	// 10s ago
	assert.Equal(t, "10s ago", RelativeTimeStr(now.Add(-10*time.Second)))
	// 45 ago
	assert.Equal(t, "45m ago", RelativeTimeStr(now.Add(-45*time.Minute)))
	// 1h ago
	assert.Equal(t, "1h ago", RelativeTimeStr(now.Add(-1*time.Hour)))
}

func TestParseIsoTime(t *testing.T) {
	isoTime, err := ParseIsoTime("2023-01-27T14:40:53.858Z")
	assert.Nil(t, err)
	assert.Equal(t, 2023, isoTime.Year())
	assert.Equal(t, time.January, isoTime.Month())
	assert.Equal(t, 27, isoTime.Day())
	assert.Equal(t, 14, isoTime.Hour())
	assert.Equal(t, 40, isoTime.Minute())
	assert.Equal(t, 53, isoTime.Second())
	assert.Equal(t, 858000000, isoTime.Nanosecond())

	// This is what our GO api gives us
	isoTime, err = ParseIsoTime("2023-01-27T14:45:59.042046464Z")
	assert.Nil(t, err)

	// Invalid
	isoTime, err = ParseIsoTime("2023-01-27T14:45:59.042046464")
	assert.NotNil(t, err)
}

func TestTimeToIsoTime(t *testing.T) {
	ts := time.Date(2023, time.January, 27, 14, 45, 59, 42046464, time.UTC)
	assert.Equal(t, "2023-01-27T14:45:59.042046464Z", TimeToIsoString(ts))
}

func TestSecondsSinceEpochToTime(t *testing.T) {
	s := 1678470517
	asTime := SecondsSinceEpochToTime(int64(s))
	assert.Equal(t, 2023, asTime.Year())
	assert.Equal(t, time.March, asTime.Month())
	assert.Equal(t, 10, asTime.Day())
	assert.Equal(t, 17, asTime.Hour())
	assert.Equal(t, 48, asTime.Minute())
	assert.Equal(t, 37, asTime.Second())
}
