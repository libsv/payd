package http //nolint:dupl // simple service

import (
	"net/http"

	"github.com/labstack/echo/v4"
	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
)

type owners struct {
	svc gopayd.OwnerService
}

// NewOwnersHandler returns a new handler for owner endpoints.
func NewOwnersHandler(svc gopayd.OwnerService) *owners {
	return &owners{
		svc: svc,
	}
}

func (o *owners) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1Owner, o.owner)
}

// owner godoc
// @Summary Owner
// @Description Returns information about the wallet owner
// @Tags Owner
// @Accept json
// @Produce json
// @Success 200
// @Router /owner [GET].
func (o *owners) owner(c echo.Context) error {
	owner, err := o.svc.Owner(c.Request().Context())
	if err != nil {
		return errors.WithStack(err)
	}

	return c.JSON(http.StatusOK, owner)
}
