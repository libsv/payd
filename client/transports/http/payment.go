package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/payd/client"
	"github.com/pkg/errors"
)

type payment struct {
	svc client.PaymentService
}

// NewPayment returns a new payment.
func NewPayment(svc client.PaymentService) *payment {
	return &payment{svc: svc}
}

func (p *payment) RegisterRoutes(g *echo.Group) {
	g.POST(RouteCreatePayment, p.create)
}

func (p *payment) create(e echo.Context) error {
	var req client.CreatePayment
	if err := e.Bind(&req); err != nil {
		return errors.Wrap(err, "failed to parse payment request")
	}
	pa, err := p.svc.CreatePayment(e.Request().Context(), req)
	if err != nil {
		return err
	}
	return e.JSON(http.StatusCreated, pa)
}
