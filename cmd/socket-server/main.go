package main

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/libsv/payd/cmd/internal"
	"github.com/libsv/payd/config/databases"
	"github.com/libsv/payd/docs"
	_ "github.com/libsv/payd/docs"
	paydMiddleware "github.com/libsv/payd/transports/http/middleware"
	socMiddleware "github.com/libsv/payd/transports/sockets/middleware"

	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/theflyingcodr/sockets/client"
	smw "github.com/theflyingcodr/sockets/middleware"
	"github.com/theflyingcodr/sockets/server"

	"github.com/libsv/payd/config"
	thttp "github.com/libsv/payd/transports/http"
	tsoc "github.com/libsv/payd/transports/sockets"
)

const appname = "payd-sockets"
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
	println("socket server")
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
		WithMapi().
		WithSocket()
	// validate the config, fail if it fails.
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	config.SetupLog(cfg.Logging)
	log.Infof("\n------Environment: %s -----\n", cfg.Server)

	// setup db
	db, err := databases.NewDbSetup().SetupDb(cfg.Db)
	if err != nil {
		log.Fatalf("failed to setup database: %s", err)
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
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.HTTPErrorHandler = paydMiddleware.ErrorHandler
	p := prometheus.NewPrometheus("paydsockets", nil)
	p.Use(e)
	if cfg.Deployment.IsDev() {
		internal.PrintDev(e)
	}
	// create socket server
	s := server.New()
	defer s.Close()
	// this is our websocket endpoint, clients will hit this with the channelID they wish to connect to
	e.GET("ws/:channelID", wsHandler(s))
	// add middleware, with panic going first
	metricsMW := smw.Metrics()
	s.WithMiddleware(smw.PanicHandler, smw.Timeout(smw.NewTimeoutConfig()), metricsMW)

	// socket client server
	c := client.New(client.WithMaxMessageSize(10000))
	defer c.Close()
	c.WithMiddleware(smw.PanicHandler,
		smw.Timeout(smw.NewTimeoutConfig()),
		metricsMW,
		smw.Logger(smw.NewLoggerConfig()),
		socMiddleware.IgnoreMyMessages(cfg.Socket),
		socMiddleware.WithAppIDPayD())

	services := internal.SetupSocketDeps(cfg, db, c)

	// client handlers
	tsoc.NewPaymentRequest(services.PaymentRequestService, services.EnvelopeService).RegisterListeners(c)
	tsoc.NewPayments(services.PaymentService).RegisterListeners(c)
	tsoc.NewProofs(services.ProofService).RegisterListeners(c)

	// rest handlers
	thttp.NewPayHandler(services.PayService).RegisterRoutes(g)
	thttp.NewConnect(services.ConnectService).RegisterRoutes(g)
	thttp.NewInvoice(services.InvoiceService).RegisterRoutes(g)
	thttp.NewOwnersHandler(services.OwnerService).RegisterRoutes(g)
	thttp.NewBalance(services.BalanceService).RegisterRoutes(g)

	e.Logger.Fatal(e.Start(cfg.Server.Port))
}

func wsHandler(svr *server.SocketServer) echo.HandlerFunc {
	upgrader := websocket.Upgrader{}
	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		defer func() {
			_ = ws.Close()
		}()
		if err := svr.Listen(ws, c.Param("channelID")); err != nil {
			return err
		}

		fmt.Println("exiting listener")
		return nil
	}
}
