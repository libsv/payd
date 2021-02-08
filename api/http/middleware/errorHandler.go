package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// ErrorHandler we can flesh this out.
func ErrorHandler(err error, c echo.Context) {
	log.Error(err)
}
