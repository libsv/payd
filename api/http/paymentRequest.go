package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	gopayd "github.com/libsv/go-payd"
	"github.com/libsv/go-payd/config"
)

// paymentHandler is an http handler that supports BIP-270 requests.
type paymentHandler struct {
	svc gopayd.PaymentRequestService
	cfg *config.Paymail
	env *config.Server
}

// NewPaymentHandler will create and return a new PaymentHandler.
func NewPaymentHandler(cfg *config.Paymail, env *config.Server, svc gopayd.PaymentRequestService) *paymentHandler {
	return &paymentHandler{
		svc: svc,
		cfg: cfg,
		env: env,
	}
}

func (h *paymentHandler) RegisterRoutes(g *echo.Group) {
	g.GET(RoutePaymentRequest, h.createPaymentRequest)
	g.POST(RoutePayment, h.createPayment)
}

func (h *paymentHandler) createPaymentRequest(e echo.Context) error {
	args := gopayd.PaymentRequestArgs{
		UsePaymail: h.cfg.UsePaymail,
		Hostname:   h.env.Hostname,
	}
	pID := e.QueryParam("paymentID")
	if pID != "" {
		args.PaymentID = &pID
	}
	resp, err := h.svc.CreatePaymentRequest(e.Request().Context(), args)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, resp)
}

func (h *paymentHandler) createPayment(e echo.Context) error {
	args := gopayd.CreatePaymentArgs{}
	e.Param("paymentID")
	pID := e.Param("paymentID")
	if pID != "" {
		args.PaymentID = &pID
	}
	return nil
}
