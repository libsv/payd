package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/pkg/errors"
)

type paymentRequest struct {
	svc   payd.PaymentRequestService
	p4Cfg *config.P4
}

// NewPaymentRequests returns a new handler for payment request endpoint.
func NewPaymentRequests(svc payd.PaymentRequestService, p4Cfg *config.P4) *paymentRequest {
	return &paymentRequest{svc: svc, p4Cfg: p4Cfg}
}

func (p *paymentRequest) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1Payment, p.get)
}

// buildPaymentRequest will setup and return a new payment request.
// @Summary Request to pay an invoice and receive back outputs to use when constructing the payment transaction
// @Description Creates a payment request based on a payment id (the identifier for an invoice).
// @Tags Receive
// @Accept json
// @Produce json
// @Param paymentID path string true "Payment ID"
// @Success 201 {object} payd.PaymentRequestResponse "contains outputs, merchant data and expiry information, used by the payee to construct a transaction"
// @Failure 404 {object} payd.ClientError "returned if the paymentID has not been found"
// @Failure 400 {object} payd.ClientError "returned if the user input is invalid, usually an issue with the paymentID"
// @Failure 500 {string} string "returned if there is an unexpected internal error"
// @Router /api/v1/payment/{paymentID} [GET].
func (p *paymentRequest) get(c echo.Context) error {
	var args payd.PaymentRequestArgs
	if err := c.Bind(&args); err != nil {
		return errors.Wrap(err, "failed to bind request")
	}
	req, err := p.svc.PaymentRequest(c.Request().Context(), args)
	if err != nil {
		return errors.WithStack(err)
	}
	req.PaymentURL = fmt.Sprintf("%s/api/v1/payment/%s", p.p4Cfg.ServerHost, args.InvoiceID)

	return c.JSON(http.StatusOK, req)
}
