package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/libsv/go-payd/api/http"
	paydMiddleware "github.com/libsv/go-payd/api/http/middleware"
	"github.com/libsv/go-payd/config"
	"github.com/libsv/go-payd/service"
)

const appname = "payd"

func main() {
	cfg := config.NewViperConfig(appname).
		WithServer().
		WithDb().
		WithDeployment(appname).
		WithLog().
		WithPaymail()
	config.SetupLog(cfg.Logging)
	log.Infof("\n------Environment: %s -----\n", cfg.Server)
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = paydMiddleware.ErrorHandler
	g := e.Group("/v1")
	http.NewPaymentHandler(cfg.Paymail, service.NewPaymentService()).RegisterRoutes(g)
	e.Logger.Fatal(e.Start(cfg.Server.Hostname))
}
