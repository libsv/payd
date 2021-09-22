package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
)

type owners struct {
	svc gopayd.OwnerService
}

func NewOwnersHandler(svc gopayd.OwnerService) *owners {
	return &owners{
		svc: svc,
	}
}

func (o *owners) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1Owner, o.owner)
}

func (o *owners) owner(c echo.Context) error {
	owner, err := o.svc.Owner(c.Request().Context())
	if err != nil {
		return errors.WithStack(err)
	}

	return c.JSON(http.StatusOK, owner)
}
