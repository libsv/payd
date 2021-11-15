package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/payd/log"
	validator "github.com/theflyingcodr/govalidator"
	"github.com/theflyingcodr/lathos"
	"github.com/theflyingcodr/lathos/errs"
)

// ErrorHandler we can flesh this out.
func ErrorHandler(l log.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if err == nil {
			return
		}
		var valErr validator.ErrValidation
		if errors.As(err, &valErr) {
			resp := map[string]interface{}{
				"errors": valErr,
			}
			_ = c.JSON(http.StatusBadRequest, resp)
			return
		}

		type errResp struct {
			ID      string `json:"id"`
			Code    string `json:"code"`
			Title   string `json:"title"`
			Message string `json:"message"`
		}

		// Internal error, log it to a system and convert the error to an internal err.
		if !lathos.IsClientError(err) {
			internalErr := errs.NewErrInternal(err, nil)
			l.Error(internalErr, "internal error")

			_ = c.JSON(http.StatusInternalServerError, errResp{
				ID:      internalErr.ID(),
				Title:   "Internal Server Error",
				Code:    "500",
				Message: internalErr.Message(),
			})
			return
		}
		var clientErr lathos.ClientError
		errors.As(err, &clientErr)
		resp := errResp{
			ID:      clientErr.ID(),
			Code:    clientErr.Code(),
			Title:   clientErr.Title(),
			Message: clientErr.Detail(),
		}
		if lathos.IsNotFound(err) {
			_ = c.JSON(http.StatusNotFound, resp)
			return
		}
		if lathos.IsDuplicate(err) {
			_ = c.JSON(http.StatusConflict, resp)
			return
		}
		if lathos.IsNotAuthenticated(err) {
			_ = c.JSON(http.StatusUnauthorized, resp)
			return
		}
		if lathos.IsNotAuthorised(err) {
			_ = c.JSON(http.StatusForbidden, resp)
			return
		}
		if lathos.IsCannotProcess(err) {
			_ = c.JSON(http.StatusUnprocessableEntity, resp)
			return
		}
	}
}
