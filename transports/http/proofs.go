package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/go-bk/envelope"

	gopayd "github.com/libsv/payd"
)

// proofs is used to accept merkle proofs from transactions
// submitted by the payment protocol server.
type proofs struct {
	svc gopayd.ProofsService
}

// NewProofs will setup and return a new proofs http handler.
func NewProofs(svc gopayd.ProofsService) *proofs {
	return &proofs{svc: svc}
}

// RegisterRoutes will setup all proof routes with the supplied echo group.
func (p *proofs) RegisterRoutes(g *echo.Group) {
	g.POST(RouteProofs, p.create)
}

func (p *proofs) create(c echo.Context) error {
	var req envelope.JSONEnvelope
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := p.svc.Create(c.Request().Context(), req); err != nil {
		return nil
	}
	return c.NoContent(http.StatusCreated)
}
