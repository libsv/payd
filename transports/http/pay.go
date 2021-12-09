package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/payd"
	"github.com/pkg/errors"
)

type pay struct {
	svc payd.PayService
}

// NewPayHandler returns a new handler for pay endpoints.
func NewPayHandler(svc payd.PayService) *pay {
	return &pay{
		svc: svc,
	}
}

// RegisterRoutes registers the pay routes.
func (p *pay) RegisterRoutes(g *echo.Group) {
	g.POST(RouteV1Pay, p.pay)
}

// pay will send a payment to a provided url
// @Summary Make a payment
// @Tags Pay
// @Accept json
// @Produce json
// @Param body body payd.PayRequest true "Pay to url"
// @Success 201
// @Router /api/v1/pay [POST].
func (p *pay) pay(c echo.Context) error {
	var req payd.PayRequest
	if err := c.Bind(&req); err != nil {
		return errors.Wrap(err, "failed to process payment request")
	}
	resp, err := p.svc.Pay(c.Request().Context(), req)
	if err != nil {
		return errors.WithStack(err)
	}
	return c.JSON(http.StatusCreated, resp)
}
