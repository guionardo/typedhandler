package typedhandler

import (
	"errors"
	"sync"
	"time"
)

var (
	timeLayouts = []string{
		time.DateTime,
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC1123,
		time.RFC1123Z,
		time.ANSIC,
		time.DateOnly,
		time.TimeOnly,
	}
	timeLayoutsLock  sync.RWMutex
	ErrNoTimeLayouts = errors.New("no time layouts configured for parsing")
	ErrTimeParsing   = errors.New("unable to parse time with provided layouts")
)

// SetTimeLayouts sets the time layouts used for parsing time strings.
func SetTimeLayouts(layouts []string) {
	timeLayoutsLock.Lock()
	defer timeLayoutsLock.Unlock()

	timeLayouts = layouts
}

// ParseTime attempts to parse the given string into a time.Time
// using the configured time layouts.
// It returns the parsed time.Time or an error if parsing fails.
// To avoid repeated parsing failures, it rearranges the layouts
// to prioritize successful formats in future calls.
func ParseTime(value string) (time.Time, error) {
	timeLayoutsLock.RLock()
	defer timeLayoutsLock.RUnlock()

	if len(timeLayouts) == 0 {
		return time.Time{}, ErrNoTimeLayouts
	}

	for index, layout := range timeLayouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			rearrangeLayoutToFront(index)
			return t, nil
		}
	}

	return time.Time{}, ErrTimeParsing
}

func rearrangeLayoutToFront(index int) {
	if index == 0 {
		return
	}
	timeLayouts[0], timeLayouts[index] = timeLayouts[index], timeLayouts[0]
}
