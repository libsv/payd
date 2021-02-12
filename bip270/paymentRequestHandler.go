package bip270

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	phttp "github.com/libsv/go-payd/http"
	"github.com/libsv/go-payd/ppctl"
)

// paymentRequestHandler is an http handler that supports BIP-270 requests.
type paymentRequestHandler struct {
	svc ppctl.PaymentRequestService
}

// NewPaymentRequestHandler will create and return a new PaymentRequestHandler.
func NewPaymentRequestHandler(svc ppctl.PaymentRequestService) *paymentRequestHandler {
	return &paymentRequestHandler{
		svc: svc,
	}
}

// RegisterRoutes will setup all routes with an echo group.
func (h *paymentRequestHandler) RegisterRoutes(g *echo.Group) {
	g.GET(phttp.RoutePaymentRequest, h.createPaymentRequest)
}

func (h *paymentRequestHandler) createPaymentRequest(e echo.Context) error {
	args := ppctl.PaymentRequestArgs{
		PaymentID: e.Param("paymentID"),
	}
	resp, err := h.svc.CreatePaymentRequest(e.Request().Context(), args)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, resp)
}
