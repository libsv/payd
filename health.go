package payd

import "context"

// HealthService for checking the overall health of payd.
type HealthService interface {
	Health(ctx context.Context) error
}

// HealthCheck for checking the health of a specific component.
type HealthCheck interface {
	Start() error
}
