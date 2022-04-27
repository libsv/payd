package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/libsv/payd"
	"github.com/pkg/errors"
)

type users struct {
	svc payd.UserService
}

// NewUsersHandler returns a new handler for owner endpoints.
func NewUsersHandler(svc payd.UserService) *users {
	return &users{
		svc: svc,
	}
}

// RegisterRoutes will setup the http handler with the echo group.
func (u *users) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1UserID, u.user)
	g.POST(RouteV1User, u.create)
}

// user will return information on the user associated with the id.
// @Router /user/:id [GET].
func (u *users) user(c echo.Context) error {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return errors.Wrap(err, "user_id is not a valid number")
	}
	user, err := u.svc.ReadUser(c.Request().Context(), userID)
	if err != nil {
		return errors.WithStack(err)
	}

	return c.JSON(http.StatusOK, user)
}

// user will return information on the user associated with the id.
// @Router /user/:id [GET].
func (u *users) create(e echo.Context) error {
	var req payd.CreateUserArgs
	if err := e.Bind(&req); err != nil {
		return errors.Wrap(err, "failed to bind request")
	}
	res, err := u.svc.CreateUser(e.Request().Context(), req)
	if err != nil {
		return e.JSON(http.StatusBadRequest, struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
	}

	return e.JSON(http.StatusOK, res)
}
