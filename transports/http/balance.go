package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

type balance struct {
	svc payd.BalanceService
}

// NewBalance will setup and return a balance handler.
func NewBalance(svc payd.BalanceService) *balance {
	return &balance{svc: svc}
}

// RegisterRoutes will hook up the routes to the echo group.
func (b *balance) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1Balance, b.balance)
}

// balance godoc
// @Summary Balance
// @Description Returns current balance, which is a sum of unspent txos
// @Tags Balance
// @Accept json
// @Produce json
// @Success 200
// @Router /api/v1/balance [GET].
func (b *balance) balance(e echo.Context) error {
	resp, err := b.svc.Balance(e.Request().Context())
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, resp)
}
