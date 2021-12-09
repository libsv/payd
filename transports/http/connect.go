package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

type connect struct {
	svc payd.ConnectService
}

// NewConnect will setup a connection handler.
func NewConnect(svc payd.ConnectService) *connect {
	return &connect{svc: svc}
}

func (c *connect) RegisterRoutes(e *echo.Group) {
	e.POST(RouteV1Connect, c.connect)
}

func (c *connect) connect(e echo.Context) error {
	var args payd.ConnectArgs
	if err := e.Bind(&args); err != nil {
		return err
	}
	if err := c.svc.Connect(e.Request().Context(), args); err != nil {
		return errors.WithStack(err)
	}
	return e.NoContent(http.StatusCreated)
}
