package middleware

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/libsv/payd"
	"github.com/libsv/payd/log"
	"github.com/libsv/payd/session"
	"github.com/pkg/errors"
)

// AuthUser infer user from header. Default to 1 if not supplied.
func AuthUser(l log.Logger, svc payd.UserService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			hdr := c.Request().Header.Get("x-user")
			if hdr == "" {
				hdr = "1"
			}
			uID, err := strconv.ParseUint(hdr, 10, 64)
			if err != nil {
				return errors.WithStack(err)
			}

			u, err := svc.ReadUser(c.Request().Context(), uID)
			if err != nil {
				return errors.WithStack(err)
			}
			c.SetRequest(c.Request().WithContext(session.WithUser(c.Request().Context(), u)))
			return next(c)
		}
	}
}
