package http

import (
	"github.com/labstack/echo/v4"
	"github.com/libsv/payd"
)

type pay struct {
	svc payd.PayService
}

func NewPayHandler(svc payd.PayService) *pay {
	return &pay{
		svc: svc,
	}
}

func (p *pay) RegisterRoutes(g *echo.Group) {
	g.POST(RouteV1Pay, p.pay)
}

func (p *pay) pay(c echo.Context) error {
	return nil
}
