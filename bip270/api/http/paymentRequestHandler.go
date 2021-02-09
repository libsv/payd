package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/libsv/go-payd/api/paydHttp"
	"github.com/libsv/go-payd/bip270"
	"github.com/libsv/go-payd/config"
)

// paymentRequestHandler is an http handler that supports BIP-270 requests.
type paymentRequestHandler struct {
	svc bip270.PaymentRequestService
	cfg *config.Paymail
	env *config.Server
}

// NewPaymentRequestHandler will create and return a new PaymentRequestHandler.
func NewPaymentRequestHandler(cfg *config.Paymail, env *config.Server, svc bip270.PaymentRequestService) *paymentRequestHandler {
	return &paymentRequestHandler{
		svc: svc,
		cfg: cfg,
		env: env,
	}
}

// RegisterRoutes will setup all routes with an echo group.
func (h *paymentRequestHandler) RegisterRoutes(g *echo.Group) {
	g.GET(paydHttp.RoutePaymentRequest, h.createPaymentRequest)
}

func (h *paymentRequestHandler) createPaymentRequest(e echo.Context) error {
	args := bip270.PaymentRequestArgs{
		UsePaymail: h.cfg.UsePaymail,
		Hostname:   h.env.Hostname,
		PaymentID:  e.QueryParam("paymentID"),
	}
	resp, err := h.svc.CreatePaymentRequest(e.Request().Context(), args)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, resp)
}
