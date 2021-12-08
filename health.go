package payd

import "context"

type HealthService interface {
	Health(ctx context.Context) error
}
