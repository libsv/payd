package main

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/libsv/go-bc/spv"
	"github.com/libsv/payd/client/data/ppctl"
	"github.com/libsv/payd/client/data/regtest"
	"github.com/libsv/payd/client/data/spvstore"
	"github.com/libsv/payd/client/service"
	chttp "github.com/libsv/payd/client/transports/http"
	"github.com/libsv/payd/config"
	"github.com/libsv/payd/config/databases"
	paydSQL "github.com/libsv/payd/data/sqlite"
	pservice "github.com/libsv/payd/service"
	paydMiddleware "github.com/libsv/payd/transports/http/middleware"
)

const appname = "spvclient"

func main() {
	cfg := config.NewViperConfig(appname).
		WithServer().
		WithDb().
		WithDeployment(appname).
		WithLog().
		WithRegtest().
		WithPpctl()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	db, err := databases.NewDbSetup().SetupDb(cfg.Db)
	if err != nil {
		log.Fatalf("failed to setup database: %s", err)
	}
	defer func() {
		_ = db.Close()
	}()

	e := echo.New()
	e.HideBanner = true
	g := e.Group("/")

	e.Use(
		middleware.Recover(),
		middleware.Logger(),
		middleware.RequestID(),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		}),
	)
	e.HTTPErrorHandler = paydMiddleware.ErrorHandler

	sqlLiteStore := paydSQL.NewSQLiteStore(db)

	pkSvc := pservice.NewPrivateKeys(sqlLiteStore, false)

	rt := regtest.NewRegtest(cfg.Regtest, &http.Client{Timeout: 30 * time.Second})

	var txStore spv.TxStore = spvstore.NewSPVStore(rt)
	var mpStore spv.MerkleProofStore = spvstore.NewSPVStore(rt)

	fSvc := service.NewFundService(rt, sqlLiteStore, pkSvc)

	spv, err := spv.NewEnvelopeCreator(txStore, mpStore)
	if err != nil {
		log.Fatal(err)
	}

	ppctlSvc := ppctl.NewPPCTL(&http.Client{Timeout: 30 * time.Second}, cfg.Ppctl)

	chttp.NewPayment(service.NewPayment(spv, ppctlSvc, fSvc, pkSvc, sqlLiteStore)).RegisterRoutes(g)
	chttp.NewFund(fSvc).RegisterRoutes(g)
	chttp.NewTxStatus(service.NewTxStatusService(ppctlSvc)).RegisterRoutes(g)

	e.Logger.Fatal(e.Start(cfg.Server.Port))
}
