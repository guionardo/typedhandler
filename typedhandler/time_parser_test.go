package typedhandler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTime(t *testing.T) { //nolint
	SetTimeLayouts([]string{time.RFC3339, time.RFC1123})

	// Test RFC3339 format
	input := "Mon, 02 Jan 2006 15:04:05 MST"
	expected := time.Date(2006, time.January, 2, 15, 4, 5, 0, time.FixedZone("MST", 0))

	parsedTime, err := ParseTime(input)
	require.NoError(t, err)
	assert.Equal(t, expected, parsedTime)
	assert.Equal(t, time.RFC1123, timeLayouts[0])

	// Test invalid time string
	_, err = ParseTime("invalid-time-string")
	require.ErrorIs(t, err, ErrTimeParsing)

	// Test with no layouts configured
	SetTimeLayouts([]string{})

	_, err = ParseTime("2023-10-05T14:48:00Z")
	require.ErrorIs(t, err, ErrNoTimeLayouts)

	SetTimeLayouts([]string{time.RFC3339, time.RFC1123})
}
