package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/libsv/go-bc/spv"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/tonicpow/go-minercraft"

	"github.com/libsv/payd/data/mapi"
	_ "github.com/libsv/payd/docs"

	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"

	"github.com/libsv/payd/config/databases"
	paydSQL "github.com/libsv/payd/data/sqlite"
	"github.com/libsv/payd/service"
	thttp "github.com/libsv/payd/transports/http"

	"github.com/libsv/payd/config"
	dataHttp "github.com/libsv/payd/data/http"
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
	println("\033[32m" + banner + "\033[0m")
	cfg := config.NewViperConfig(appname).
		WithServer().
		WithDb().
		WithDeployment(appname).
		WithLog().
		WithHeadersClient().
		WithPaymail().
		WithWallet().
		WithMapi()
	// validate the config, fail if it fails.
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	config.SetupLog(cfg.Logging)
	log.Infof("\n------Environment: %s -----\n", cfg.Server)
	db, err := databases.NewDbSetup().SetupDb(cfg.Db)
	if err != nil {
		log.Fatalf("failed to setup database: %s", err)
	}
	// nolint:errcheck // dont care about error.
	defer db.Close()

	e := echo.New()
	e.HideBanner = true
	g := e.Group("/")
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.HTTPErrorHandler = paydMiddleware.ErrorHandler

	// setup stores
	mapiCli, err := minercraft.NewClient(nil, nil, []*minercraft.Miner{
		{
			Name:  cfg.Mapi.MinerName,
			Token: cfg.Mapi.Token,
			URL:   cfg.Mapi.URL,
		},
	})
	if err != nil {
		log.Fatal(mapiCli)
	}
	sqlLiteStore := paydSQL.NewSQLiteStore(db)
	mapiStore := mapi.NewMapi(cfg.Mapi, cfg.Server, mapiCli)
	spvv, err := spv.NewPaymentVerifier(dataHttp.NewHeaderSVConnection(&http.Client{Timeout: time.Duration(cfg.HeadersClient.Timeout) * time.Second}, cfg.HeadersClient.Address))
	if err != nil {
		log.Fatalf("failed to create spv client %w", err)
	}

	// spvb, err := spv.NewEnvelopeCreator(nil, )

	// setup services
	privKeySvc := service.NewPrivateKeys(sqlLiteStore, cfg.Wallet.Network == "mainnet")
	destSvc := service.NewDestinationsService(privKeySvc, sqlLiteStore, sqlLiteStore, mapiStore)
	paymentSvc := service.NewPayments(spvv, sqlLiteStore, sqlLiteStore, sqlLiteStore, &paydSQL.Transacter{}, mapiStore, sqlLiteStore)

	thttp.NewInvoice(service.NewInvoice(cfg.Server, sqlLiteStore, destSvc, &paydSQL.Transacter{})).
		RegisterRoutes(g)
	thttp.NewBalance(service.NewBalance(sqlLiteStore)).RegisterRoutes(g)
	thttp.NewProofs(service.NewProofsService(sqlLiteStore)).RegisterRoutes(g)
	thttp.NewDestinations(destSvc).RegisterRoutes(g)
	thttp.NewPayments(paymentSvc).RegisterRoutes(g)
	thttp.NewOwnersHandler(service.NewOwnerService(sqlLiteStore)).RegisterRoutes(g)
	thttp.NewPayHandler(service.NewPayService(sqlLiteStore, dataHttp.NewP4(&http.Client{}), privKeySvc, nil)).RegisterRoutes(g)

	if cfg.Deployment.IsDev() {
		printDev(e)
	}
	e.Logger.Fatal(e.Start(cfg.Server.Port))
}

// printDev outputs some useful dev information such as http routes
// and current settings being used.
func printDev(e *echo.Echo) {
	fmt.Println("==================================")
	fmt.Println("DEV mode, printing http routes:")
	for _, r := range e.Routes() {
		fmt.Printf("%s: %s\n", r.Method, r.Path)
	}
	fmt.Println("==================================")
	fmt.Println("DEV mode, printing settings:")
	for _, v := range viper.AllKeys() {
		fmt.Printf("%s: %v\n", v, viper.Get(v))
	}
	fmt.Println("==================================")
}
