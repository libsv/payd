package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/libsv/payd/cmd/internal"
	"github.com/libsv/payd/docs"

	"github.com/libsv/payd/config/databases"
	_ "github.com/libsv/payd/docs"
	"github.com/libsv/payd/log"
	thttp "github.com/libsv/payd/transports/http"

	"github.com/libsv/payd/config"
	paydMiddleware "github.com/libsv/payd/transports/http/middleware"
)

const appname = "payd"
const banner = `
====================================================================
         _               _           _        _          _         
        /\ \            / /\        /\ \     /\_\       /\ \       
       /  \ \          / /  \       \ \ \   / / /      /  \ \____  
      / /\ \ \        / / /\ \       \ \ \_/ / /      / /\ \_____\ 
     / / /\ \_\      / / /\ \ \       \ \___/ /      / / /\/___  / 
    / / /_/ / /     / / /  \ \ \       \ \ \_/      / / /   / / /  
   / / /__\/ /     / / /___/ /\ \       \ \ \      / / /   / / /   
  / / /_____/     / / /_____/ /\ \       \ \ \    / / /   / / /    
 / / /           / /_________/\ \ \       \ \ \   \ \ \__/ / /     
/ / /           / / /_       __\ \_\       \ \_\   \ \___\/ /      
\/_/            \_\___\     /____/_/        \/_/    \/_____/  
====================================================================
`

// create godoc
// @title Payd
// @version 0.0.1
// @description Payd is a txo and key manager, with a common interface that can be implemented by wallets.
// @termsOfService https://github.com/libsv/payd/blob/master/CODE_OF_CONDUCT.md
// @license.name ISC
// @license.url https://github.com/libsv/payd/blob/master/LICENSE
// @host localhost:8443
// @BasePath /api
// @schemes:
//	- http
//	- https
func main() {
	println("rest server")
	println("\033[32m" + banner + "\033[0m")
	config.SetupDefaults()
	cfg := config.NewViperConfig(appname).
		WithServer().
		WithDb().
		WithDeployment(appname).
		WithLog().
		WithHeadersClient().
		WithWallet().
		WithP4().
		WithMapi()
	log := log.NewZero(cfg.Logging)
	// validate the config, fail if it fails.
	if err := cfg.Validate(); err != nil {
		log.Fatal(err, "validation errors")
	}
	log.Infof("------Environment: %v -----", cfg.Server)
	db, err := databases.NewDbSetup().SetupDb(log, cfg.Db)
	if err != nil {
		log.Fatal(err, "failed to setup database")
	}
	// nolint:errcheck // dont care about error.
	defer db.Close()

	e := echo.New()
	e.HideBanner = true
	g := e.Group("/")
	if cfg.Server.SwaggerEnabled {
		docs.SwaggerInfo.Host = cfg.Server.SwaggerHost
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.HTTPErrorHandler = paydMiddleware.ErrorHandler(log)

	// setup deps
	services := internal.SetupRestDeps(cfg, log, db)

	thttp.NewInvoice(services.InvoiceService).
		RegisterRoutes(g)
	thttp.NewBalance(services.BalanceService).RegisterRoutes(g)
	thttp.NewProofs(services.ProofService).RegisterRoutes(g)
	thttp.NewDestinations(services.DestinationService).RegisterRoutes(g)
	thttp.NewPayments(services.PaymentService).RegisterRoutes(g)
	thttp.NewOwnersHandler(services.OwnerService).RegisterRoutes(g)
	thttp.NewPayHandler(services.PayService).RegisterRoutes(g)

	if cfg.Deployment.IsDev() {
		internal.PrintDev(e)
	}
	e.Logger.Fatal(e.Start(cfg.Server.Port))
}
