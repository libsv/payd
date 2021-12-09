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
// @Description Given an paymentID, and an spvEnvelope, we will validate the payment and inputs used are valid and that it covers the payment.
// @Tags Receive
// @Accept json
// @Produce json
// @Param paymentID path string true "Payment ID"
// @Failure 400 {object} payd.ClientError "returned if the paymentID is empty or payment isn't valid"
// @Failure 404 {object} payd.ClientError "returned if the paymentID has not been found"
// @Success 200
// @Router /api/v1/payments/{paymentID} [POST].
func (p *payments) create(e echo.Context) error {
	var req p4.Payment
	if err := e.Bind(&req); err != nil {
		return errors.Wrap(err, "failed to bind request")
	}
	if err := p.svc.PaymentCreate(e.Request().Context(), payd.PaymentCreateArgs{PaymentID: e.Param("paymentID")}, req); err != nil {
		return errors.WithStack(err)
	}
	return e.NoContent(http.StatusNoContent)
}
