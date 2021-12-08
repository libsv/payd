package payd

import "context"

// HealthService for checking health.
type HealthService interface {
	Health(ctx context.Context) error
}
