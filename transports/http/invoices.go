package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

type invoices struct {
	svc payd.InvoiceService
}

// NewInvoice will setup and return a new invoices handler.
func NewInvoice(svc payd.InvoiceService) *invoices {
	return &invoices{svc: svc}
}

// RegisterRoutes will hook up the routes to the echo group.
func (i *invoices) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1Invoices, i.invoices)
	g.GET(RouteV1Invoice, i.invoice)
	g.POST(RouteV1Invoices, i.create)
	g.DELETE(RouteV1Invoice, i.delete)
}

// invoices godoc
// @Summary Invoices
// @Description Returns all invoices currently stored
// @Tags Invoices
// @Accept json
// @Produce json
// @Success 200
// @Router /invoices [GET].
func (i *invoices) invoices(e echo.Context) error {
	ii, err := i.svc.Invoices(e.Request().Context())
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, ii)
}

// invoices godoc
// @Summary Invoices
// @Description Returns invoices by invoices id if exists
// @Tags Invoices
// @Accept json
// @Produce json
// @Param invoiceID path string true "Invoice ID"
// @Success 200
// @Router /invoices/{invoiceID} [GET].
func (i *invoices) invoice(e echo.Context) error {
	var args payd.InvoiceArgs
	if err := e.Bind(&args); err != nil {
		return errors.Wrap(err, "failed to parse invoices args")
	}
	inv, err := i.svc.Invoice(e.Request().Context(), args)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, inv)
}

// create will validate and persist a new invoices.
// @Summary InvoiceCreate invoices
// @Description Creates an invoices with invoiceID and satoshis
// @Tags Invoices
// @Accept json
// @Produce json
// @Param body body payd.InvoiceCreate true "Reference and Satoshis"
// @Success 201
// @Router /invoices [POST].
func (i *invoices) create(e echo.Context) error {
	var req payd.InvoiceCreate
	if err := e.Bind(&req); err != nil {
		return errors.Wrap(err, "failed to parse invoices create req")
	}
	inv, err := i.svc.Create(e.Request().Context(), req)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusCreated, inv)
}

// delete godoc
// @Summary InvoiceDelete invoices
// @Description InvoiceDelete
// @Tags Invoices
// @Accept json
// @Produce json
// @Param invoiceID path string true "invoiceID we want to remove"
// @Success 204
// @Failure 404 {object} payd.ClientError "returned if the paymentID has not been found"
// @Router /v1/invoices/{invoiceID} [DELETE].
func (i *invoices) delete(e echo.Context) error {
	var args payd.InvoiceArgs
	if err := e.Bind(&args); err != nil {
		return errors.Wrap(err, "failed to parse invoices delete args")
	}
	if err := i.svc.Delete(e.Request().Context(), args); err != nil {
		return errors.WithStack(err)
	}
	return e.NoContent(http.StatusNoContent)
}
