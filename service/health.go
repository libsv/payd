package service

import (
	"context"

	"github.com/InVisionApp/go-health/v2"
	"github.com/libsv/payd"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type healthSvc struct {
	h health.IHealth
}

// NewHealthService (NHS) will setup and return a new health service.
func NewHealthService(h health.IHealth) payd.HealthService {
	return &healthSvc{
		h: h,
	}
}

// Health will return an error if the application is in an unhealthy state.
func (h *healthSvc) Health(ctx context.Context) error {
	status, failed, err := h.h.State()
	if err != nil {
		return errors.Wrap(err, "failed to check health state")
	}
	if len(status) == 0 {
		return nil
	}
	if failed {
		log.Error().Interface("key check failed", status)
		return errors.New("all healthchecks failed")
	}

	return nil
}
