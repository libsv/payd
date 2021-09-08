package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/payd/client"
)

type txstatus struct {
	svc client.TxStatusService
}

// NewTxStatus returns a new txstatus.
func NewTxStatus(svc client.TxStatusService) *txstatus {
	return &txstatus{svc: svc}
}

func (t *txstatus) RegisterRoutes(g *echo.Group) {
	g.GET(RouteTxStatus, t.status)
}

func (t *txstatus) status(c echo.Context) error {
	args := client.TxStatusArgs{
		TxID: c.Param("txid"),
	}
	txStatus, err := t.svc.Status(c.Request().Context(), args)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, txStatus)
}
