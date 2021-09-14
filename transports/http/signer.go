package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	gopayd "github.com/libsv/payd"
)

type signer struct {
	svc gopayd.SignerService
}

func NewSignerHandler(svc gopayd.SignerService) *signer {
	return &signer{
		svc: svc,
	}
}

func (s *signer) RegisterRoutes(g *echo.Group) {
	g.POST(RouteFundAndSign, s.sign)
}

func (s *signer) sign(c echo.Context) error {
	var req gopayd.FundAndSignTxRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	resp, err := s.svc.FundAndSignTx(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
