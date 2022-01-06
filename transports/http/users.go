package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/payd"
	"github.com/pkg/errors"
)

type users struct {
	svc payd.UserService
}

// NewUsersHandler returns a new handler for owner endpoints.
func NewUsersHandler(svc payd.OwnerService) *owners {
	return &owners{
		svc: svc,
	}
}

// RegisterRoutes will setup the http handler with the echo group.
func (u *users) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1UserByHandle, u.user)
}

// user will return information on the user associated with the handle.
// @Router /v1/user/:handle [GET].
func (u *users) user(c echo.Context) error {
	user, err := u.svc.Read(c.Request().Context(), c.Param("handle"))
	if err != nil {
		return errors.WithStack(err)
	}

	return c.JSON(http.StatusOK, user)
}
