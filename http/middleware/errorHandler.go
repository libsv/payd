package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// ErrorHandler we can flesh this out.
func ErrorHandler(err error, c echo.Context) {
	if err != nil {

		log.Error(err)
		c.String(http.StatusBadRequest, err.Error())
	}
}
