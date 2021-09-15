package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	gopayd "github.com/libsv/payd"
)

type fund struct {
	svc gopayd.FundService
}

// NewFundHandler returns a new funds handler.
func NewFundHandler(svc gopayd.FundService) *fund {
	return &fund{
		svc: svc,
	}
}

func (f *fund) RegisterRoutes(g *echo.Group) {
	g.GET(RouteFundGet, f.get)
	g.POST(RouteFundRequestAmount, f.requestAmount)
	g.POST(RouteFundAdd, f.add)
	g.PUT(RouteFundSpend, f.spend)
}

func (f *fund) add(c echo.Context) error {
	var req gopayd.FundAddRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	fund, err := f.svc.FundsAdd(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, fund)
}

func (f *fund) requestAmount(c echo.Context) error {
	var req gopayd.FundsRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	amount, err := strconv.ParseUint(c.Param("amount"), 10, 64)
	if err != nil {
		return err
	}

	args := gopayd.FundsGetArgs{
		Account: c.Request().Header.Get("x-account"),
		Amount:  amount,
	}

	funds, err := f.svc.FundsGetAmount(c.Request().Context(), req, args)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, funds)
}

func (f *fund) get(c echo.Context) error {
	args := gopayd.FundsGetArgs{
		Account: c.Request().Header.Get("x-account"),
	}

	funds, err := f.svc.FundsGet(c.Request().Context(), args)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, funds)
}

func (f *fund) spend(c echo.Context) error {
	var req gopayd.FundsSpendReq
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := f.svc.FundsSpend(c.Request().Context(), req, gopayd.FundsSpendArgs{
		Account: c.Request().Header.Get("x-account"),
	}); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
