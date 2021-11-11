package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

type transactions struct {
	svc payd.TransactionService
}

func NewTransactions(svc payd.TransactionService) *transactions {
	return &transactions{svc: svc}
}

func (t *transactions) RegisterRoutes(g *echo.Group) {
	g.POST(RouteV1Transaction, t.submit)
}

func (t *transactions) submit(c echo.Context) error {
	var req payd.TransactionSubmit
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := t.svc.Submit(c.Request().Context(),
		payd.TransactionSubmitArgs{InvoiceID: c.Param("invoiceID")}, req); err != nil {
		return errors.WithStack(err)
	}
	return c.NoContent(http.StatusCreated)
}
