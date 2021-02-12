package bip270

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	phttp "github.com/libsv/go-payd/http"

	"github.com/libsv/go-payd/ppctl"
)

// paymentHandler is an http handler that supports BIP-270 requests.
type paymentHandler struct {
	svc ppctl.PaymentService
}

// NewPaymentHandler will create and return a new PaymentHandler.
func NewPaymentHandler(svc ppctl.PaymentService) *paymentHandler {
	return &paymentHandler{
		svc: svc,
	}
}

// RegisterRoutes will setup all routes with an echo group.
func (h *paymentHandler) RegisterRoutes(g *echo.Group) {
	g.POST(phttp.RoutePayment, h.createPayment)
}

func (h *paymentHandler) createPayment(e echo.Context) error {
	args := ppctl.CreatePaymentArgs{
		PaymentID: e.Param("paymentID"),
	}
	var req ppctl.CreatePayment
	if err := phttp.Bind(e.Request().Body, &req); err != nil {
		return errors.WithStack(err)
	}
	resp, err := h.svc.Create(e.Request().Context(), args, req)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusCreated, resp)
}
