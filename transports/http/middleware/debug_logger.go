package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

// DebugDumpBody dumps the body of the request and response provided the
// passed in log level is set.
func DebugDumpBody(level log.Lvl) echo.MiddlewareFunc {
	return middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Skipper: func(c echo.Context) bool {
			return log.Level() > level || !strings.HasPrefix(c.Path(), "/api/v1/")
		},
		Handler: func(e echo.Context, reqBody, respBody []byte) {
			log.Debug(e.Request().URL.String())
			log.Debug("request: ", string(reqBody))
			log.Debug("response:", string(respBody))
		},
	})
}
