package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/payd"
	"github.com/pkg/errors"
)

type health struct {
	svc payd.HealthService
}

// NewHealthHandler sets up and returns a new health handler.
func NewHealthHandler(svc payd.HealthService) *health {
	return &health{
		svc: svc,
	}
}

// RegisterRoutes will hook up the routes to the echo group.
func (h *health) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1Health, h.health)
}

func (h *health) health(c echo.Context) error {
	if err := h.svc.Health(c.Request().Context()); err != nil {
		return errors.WithStack(err)
	}
	return c.NoContent(http.StatusNoContent)
}
