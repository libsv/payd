package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/payd/client"
)

type fund struct {
	svc client.FundService
}

func NewFund(svc client.FundService) *fund {
	return &fund{svc: svc}
}

func (f *fund) RegisterRoutes(g *echo.Group) {
	g.POST(RouteAddFund, f.seed)
}

func (f *fund) seed(e echo.Context) error {
	var req client.FundSeed
	if err := e.Bind(&req); err != nil {
		return err
	}
	fund, err := f.svc.Seed(e.Request().Context(), req)
	if err != nil {
		return err
	}
	return e.JSON(http.StatusCreated, fund)
}
