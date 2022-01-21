package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/libsv/go-p4"
	"github.com/libsv/payd"
)

type payments struct {
	svc payd.PaymentsService
}

// NewPayments will setup and return a new payments http handler.
func NewPayments(svc payd.PaymentsService) *payments {
	return &payments{svc: svc}
}

// RegisterRoutes will setup all proof routes with the supplied echo group.
func (p *payments) RegisterRoutes(g *echo.Group) {
	g.POST(RouteV1Payment, p.create)
}

// create will validate and store a payment if valid.
// @Summary Validate and store a payment.
// @Description Given an invoiceID, and an spvEnvelope, we will validate the payment and inputs used are valid and that it covers the invoice.
// @Tags Payments
// @Accept json
// @Produce json
// @Param invoiceID path string true "Invoice ID"
// @Failure 400 {object} payd.ClientError "returned if the invoiceID is empty or payment isn't valid"
// @Failure 404 {object} payd.ClientError "returned if the invoiceID has not been found"
// @Success 200
// @Router /v1/payments/{invoiceID} [POST].
func (p *payments) create(e echo.Context) error {
	var req p4.Payment
	if err := e.Bind(&req); err != nil {
		return errors.Wrap(err, "failed to bind request")
	}
	ack, err := p.svc.PaymentCreate(e.Request().Context(), payd.PaymentCreateArgs{InvoiceID: e.Param("invoiceID")}, req)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusCreated, ack)
}
