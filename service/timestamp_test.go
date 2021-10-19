package service_test

import (
	"testing"
	"time"

	"github.com/libsv/payd/service"
	"github.com/stretchr/testify/assert"
)

func TestTimestampService_Nanosecond(t *testing.T) {
	t.Parallel()

	svc := service.NewTimestampService()

	t.Run("returned timestamp is always now", func(t *testing.T) {
		prev := svc.Nanosecond()
		for i := 0; i < 1000; i++ {
			ts := svc.Nanosecond()
			assert.GreaterOrEqual(t, ts, prev)
			prev = ts
		}
	})
}

func TestTimestampService_NowUTC(t *testing.T) {
	t.Parallel()

	svc := service.NewTimestampService()

	t.Run("returned timestamp is always now", func(t *testing.T) {
		prev := svc.NowUTC()
		for i := 0; i < 1000; i++ {
			time.Sleep(4000 * time.Nanosecond)
			ts := svc.NowUTC()
			assert.GreaterOrEqual(t, ts.Nanosecond(), prev.Nanosecond())
			prev = ts
		}
	})
}
