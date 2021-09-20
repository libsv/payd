package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	gopayd "github.com/libsv/payd"
)

type destination struct {
	svc gopayd.DestinationService
}

func NewDestinationHandler(svc gopayd.DestinationService) *destination {
	return &destination{svc: svc}
}

func (d *destination) RegisterRoutes(g *echo.Group) {
	g.GET(RouteDestinations, d.get)
}

func (d *destination) get(c echo.Context) error {
	args := gopayd.DestinationArgs{
		PaymentID: c.Param("paymentID"),
	}

	destinations, err := d.svc.Destinations(c.Request().Context(), args)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, destinations)
}
