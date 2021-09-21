package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	gopayd "github.com/libsv/payd"
)

type invoice struct {
	svc gopayd.InvoiceService
}

// NewInvoice will setup and return a new invoice handler.
func NewInvoice(svc gopayd.InvoiceService) *invoice {
	return &invoice{svc: svc}
}

// RegisterRoutes will hook up the routes to the echo group.
func (i *invoice) RegisterRoutes(g *echo.Group) {
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
// @Router v1/invoices [GET].
func (i *invoice) invoices(e echo.Context) error {
	ii, err := i.svc.Invoices(e.Request().Context())
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, ii)
}

// invoice godoc
// @Summary Invoices
// @Description Returns invoice by invoice id if exists
// @Tags Invoices
// @Accept json
// @Produce json
// @Param invoiceID path string true "Invoice ID"
// @Success 200
// @Router v1/invoices/{invoiceID} [GET].
func (i *invoice) invoice(e echo.Context) error {
	var args gopayd.InvoiceArgs
	if err := e.Bind(&args); err != nil {
		return errors.Wrap(err, "failed to parse invoice args")
	}
	inv, err := i.svc.Invoice(e.Request().Context(), args)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, inv)
}

// create will validate and persist a new invoice.
// @Summary InvoiceCreate invoice
// @Description Creates an invoice with invoiceID and satoshis
// @Tags Invoices
// @Accept json
// @Produce json
// @Param body body gopayd.InvoiceCreate true "Reference and Satoshis"
// @Success 201
// @Router v1/invoices [POST].
func (i *invoice) create(e echo.Context) error {
	var req gopayd.InvoiceCreate
	if err := e.Bind(&req); err != nil {
		return errors.Wrap(err, "failed to parse invoice create req")
	}
	inv, err := i.svc.Create(e.Request().Context(), req)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusCreated, inv)
}

// delete godoc
// @Summary InvoiceDelete invoice
// @Description InvoiceDelete
// @Tags Invoices
// @Accept json
// @Produce json
// @Param invoiceID path string true "invoiceID we want to remove"
// @Success 200
// @Router v1/invoices/{invoiceID} [DELETE].
func (i *invoice) delete(e echo.Context) error {
	var args gopayd.InvoiceArgs
	if err := e.Bind(&args); err != nil {
		return errors.Wrap(err, "failed to parse invoice delete args")
	}
	if err := i.svc.Delete(e.Request().Context(), args); err != nil {
		return errors.WithStack(err)
	}
	return e.NoContent(http.StatusNoContent)
}
