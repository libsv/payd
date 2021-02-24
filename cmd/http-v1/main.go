package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	paydSQL "github.com/libsv/go-payd/data/sqlite"
	"github.com/libsv/go-payd/data/sqlite/schema"
	"github.com/libsv/go-payd/service"
	"github.com/libsv/go-payd/service/ppctl"
	"github.com/libsv/go-payd/transports/http"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"

	"github.com/libsv/go-payd/config"
	"github.com/libsv/go-payd/ipaymail"
	paydMiddleware "github.com/libsv/go-payd/transports/http/middleware"
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

func main() {
	println("\033[32m" + banner + "\033[0m")
	cfg := config.NewViperConfig(appname).
		WithServer().
		WithDb().
		WithDeployment(appname).
		WithLog().
		WithPaymail().
		WithWallet()
	config.SetupLog(cfg.Logging)
	log.Infof("\n------Environment: %s -----\n", cfg.Server)
	if cfg.Deployment.IsDev() {
		schema.MustSetup(cfg.Db)
	}
	db, err := sqlx.Open("sqlite3", cfg.Db.Dsn)
	if err != nil {
		log.Fatalf("failed to setup database: %s", err)
	}
	defer db.Close()

	e := echo.New()
	e.HideBanner = true
	g := e.Group("/")
	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.HTTPErrorHandler = paydMiddleware.ErrorHandler

	// setup stores
	sqlLiteStore := paydSQL.NewSQLiteStore(db)

	// setup services
	pwSvc := ppctl.NewPaymentWalletService(sqlLiteStore)
	pmSvc := ppctl.NewPaymailPaymentService(ipaymail.NewRransactionService())
	pkSvc := service.NewPrivateKeys(sqlLiteStore, cfg.Deployment.MainNet)

	http.NewPaymentHandler(
		ppctl.NewPaymentFacade(cfg.Paymail, pwSvc, pmSvc)).
		RegisterRoutes(g)
	http.NewPaymentRequestHandler(
		ppctl.NewPaymentRequestService(cfg.Server, cfg.Wallet, pkSvc, &paydSQL.SQLiteTransacter{}, sqlLiteStore)).
		RegisterRoutes(g)

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
