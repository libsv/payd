package payd

import "time"

// TimestampService for getting timestamps.
type TimestampService interface {
	Nanosecond() int
	NowUTC() time.Time
}
