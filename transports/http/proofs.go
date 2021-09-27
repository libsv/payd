package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/go-bk/envelope"
	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

// proofs is used to accept merkle proofs from transactions
// submitted by the payment protocol server.
type proofs struct {
	svc payd.ProofsService
}

// NewProofs will setup and return a new proofs http handler.
func NewProofs(svc payd.ProofsService) *proofs {
	return &proofs{svc: svc}
}

// RegisterRoutes will setup all proof routes with the supplied echo group.
func (p *proofs) RegisterRoutes(g *echo.Group) {
	g.POST(RouteV1Proofs, p.create)
}

// create godoc
// @Summary InvoiceCreate proof
// @Description Creates a json envelope proof
// @Tags Proofs
// @Accept json
// @Produce json
// @Param txid path string true "Transaction ID"
// @Param body body envelope.JSONEnvelope true "JSON Envelope"
// @Success 201
// @Router /v1/proofs/{txid} [POST].
func (p *proofs) create(c echo.Context) error {
	var req envelope.JSONEnvelope
	if err := c.Bind(&req); err != nil {
		return errors.WithStack(err)
	}
	args := payd.ProofCreateArgs{TxID: c.Param("txid")}
	if err := p.svc.Create(c.Request().Context(), args, req); err != nil {
		return errors.WithStack(err)
	}
	return c.NoContent(http.StatusCreated)
}
