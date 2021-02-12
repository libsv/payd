package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/libsv/go-payd/bip270"
	"github.com/libsv/go-payd/config"
	paydMiddleware "github.com/libsv/go-payd/http/middleware"
	"github.com/libsv/go-payd/ipaymail"
	ppctlDataNoop "github.com/libsv/go-payd/ppctl/data/noop"
	ppctlSqlite "github.com/libsv/go-payd/ppctl/data/sqlite"
	ppctlSvc "github.com/libsv/go-payd/ppctl/service"
	"github.com/libsv/go-payd/schema/sqlite"
	walletSqlite "github.com/libsv/go-payd/wallet/data/sqlite"
	walletSvc "github.com/libsv/go-payd/wallet/service"
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
	if cfg.Deployment.IsDev() {
		sqlite.MustSetup(cfg.Db)
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
	skStore := ppctlSqlite.NewScriptKey(db)
	invStore := ppctlDataNoop.NewInvoice()
	txStore := ppctlSqlite.NewTransaction(db)

	// setup services
	pwSvc := ppctlSvc.NewPaymentWalletService(skStore, invStore, txStore)
	pmSvc := ppctlSvc.NewPaymailPaymentService(ipaymail.NewRransactionService())
	pkSvc := walletSvc.NewPrivateKeys(walletSqlite.NewKeys(db), cfg.Deployment.MainNet)

	bip270.NewPaymentHandler(
		ppctlSvc.NewPaymentFacade(cfg.Paymail, pwSvc, pmSvc)).
		RegisterRoutes(g)
	bip270.NewPaymentRequestHandler(
		ppctlSvc.NewPaymentRequestService(cfg.Server, pkSvc, skStore, invStore)).
		RegisterRoutes(g)
	fmt.Println("Routes:")
	for _, r := range e.Routes() {
		fmt.Printf("%+v\n", r)
	}
	e.Logger.Fatal(e.Start(cfg.Server.Port))
}
