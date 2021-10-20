package service

import (
	"time"

	"github.com/libsv/payd"
)

type timestamp struct{}

// NewTimestampService returns a new timestamp service.
func NewTimestampService() payd.TimestampService {
	return &timestamp{}
}

// Nanosecond the current time in nanoseconds.
func (t *timestamp) Nanosecond() int {
	return time.Now().Nanosecond()
}

// NowUTC the current UTC time.
func (t *timestamp) NowUTC() time.Time {
	return time.Now().UTC()
}
