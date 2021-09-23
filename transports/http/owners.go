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

// RegisterRoutes will setup the http handler with the echo group.
func (o *owners) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1Owner, o.owner)
}

// owner will return information on the current wallet owner.
// @Summary Wallet owner information.
// @Description Returns information about the wallet owner
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} gopayd.User "Current wallet owner"
// @Router v1/owner [GET].
func (o *owners) owner(c echo.Context) error {
	owner, err := o.svc.Owner(c.Request().Context())
	if err != nil {
		return errors.WithStack(err)
	}

	return c.JSON(http.StatusOK, owner)
}
