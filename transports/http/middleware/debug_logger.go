package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func DebugDumpBody(level log.Lvl) echo.MiddlewareFunc {
	return middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Skipper: func(c echo.Context) bool {
			return log.Level() > level
		},
		Handler: func(e echo.Context, reqBody, respBody []byte) {
			log.Debug(e.Request().URL.String())
			log.Debug("request: ", string(reqBody))
			log.Debug("response:", string(respBody))
		},
	})
}
