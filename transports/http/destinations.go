package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

type destinations struct {
	svc payd.DestinationsService
}

// NewDestinations will setup and return a new destinations http handler.
func NewDestinations(svc payd.DestinationsService) *destinations {
	return &destinations{svc: svc}
}

// RegisterRoutes will hook up the routes to the echo group.
func (d *destinations) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1Destination, d.destination)
}

// destination will return destinations by invoiceID, or a 404 if none found.
// @Summary Given an invoiceID, a set of outputs and fees will be returned, if not found a 404 is returned.
// @Description Given an invoiceID, a set of outputs and fees will be returned, if not found a 404 is returned.
// @Tags Destinations
// @Accept json
// @Produce json
// @Param invoiceID path string true "Invoice ID"
// @Failure 404 {object} payd.ClientError "returned if the invoiceID has not been found"
// @Success 200
// @Router /v1/destinations/{invoiceID} [GET].
func (d *destinations) destination(e echo.Context) error {
	var args payd.DestinationsArgs
	if err := e.Bind(&args); err != nil {
		return errors.Wrap(err, "failed to parse destination args")
	}
	inv, err := d.svc.Destinations(e.Request().Context(), args)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, inv)
}
