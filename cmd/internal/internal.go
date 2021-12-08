package internal

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/libsv/payd/config"
	paydSQL "github.com/libsv/payd/data/sqlite"
	"github.com/libsv/payd/docs"
	"github.com/libsv/payd/log"
	"github.com/libsv/payd/service"
	thttp "github.com/libsv/payd/transports/http"
	paydMiddleware "github.com/libsv/payd/transports/http/middleware"
	tsoc "github.com/libsv/payd/transports/sockets"
	socMiddleware "github.com/libsv/payd/transports/sockets/middleware"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/theflyingcodr/sockets/client"
	smw "github.com/theflyingcodr/sockets/middleware"
	"github.com/theflyingcodr/sockets/server"
)

// SetupEcho will set up and return an echo server.
func SetupEcho(l log.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.HTTPErrorHandler = paydMiddleware.ErrorHandler(l)
	return e
}

// SetupSwagger will enable the swagger endpoints.
func SetupSwagger(cfg config.Server, e *echo.Echo) {
	docs.SwaggerInfo.Host = cfg.SwaggerHost
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}

// SetupHTTPEndpoints will register the http endpoints.
func SetupHTTPEndpoints(cfg config.Deployment, services *RestDeps, g *echo.Group) {
	// handlers
	thttp.NewInvoice(services.InvoiceService).
		RegisterRoutes(g)
	thttp.NewBalance(services.BalanceService).RegisterRoutes(g)
	thttp.NewProofs(services.ProofService).RegisterRoutes(g)
	thttp.NewDestinations(services.DestinationService).RegisterRoutes(g)
	thttp.NewPayments(services.PaymentService).RegisterRoutes(g)
	thttp.NewPaymentRequests(services.PaymentRequestService).RegisterRoutes(g)
	thttp.NewOwnersHandler(services.OwnerService).RegisterRoutes(g)
	thttp.NewPayHandler(services.PayService).RegisterRoutes(g)
	if cfg.Environment == "local" {
		// ugly endpoint for regtest topup - local only!
		thttp.NewTransactions(services.TransactionService).RegisterRoutes(g)
	}
}

// SetupSocketClient will setup handlers and socket server.
func SetupSocketClient(cfg config.Socket, deps *SocketDeps, c *client.Client) {
	c.WithMiddleware(smw.PanicHandler,
		smw.Timeout(smw.NewTimeoutConfig()),
		smw.Metrics(),
		smw.Logger(smw.NewLoggerConfig()),
		socMiddleware.IgnoreMyMessages(&cfg),
		socMiddleware.WithAppIDPayD()).
		WithErrorHandler(socMiddleware.ErrorHandler).
		WithServerErrorHandler(socMiddleware.ErrorMsgHandler)

	// client handlers
	tsoc.NewPaymentRequest(&paydSQL.Transacter{}, deps.PaymentRequestService, deps.EnvelopeService).
		RegisterListeners(c)
	tsoc.NewPayments(deps.PaymentService).
		RegisterListeners(c)
	tsoc.NewProofs(deps.ProofService, c).
		RegisterListeners(c)
}

// SetupSocketHTTPEndpoints will register the http endpoints used for sockets, there are some differences
// between this and the standard http endpoints in terms of data repos used.
func SetupSocketHTTPEndpoints(cfg config.Deployment, services *SocketDeps, g *echo.Group) {
	thttp.NewConnect(services.ConnectService).RegisterRoutes(g)
}

// SetupSocketServer will setup and return a socket server.
func SetupSocketServer(cfg config.Socket, e *echo.Echo) *server.SocketServer {
	// create socket server
	s := server.New()
	metricsMW := smw.Metrics()
	s.WithMiddleware(smw.PanicHandler, smw.Timeout(smw.NewTimeoutConfig()), metricsMW)

	// this is our websocket endpoint, clients will hit this with the channelID they wish to connect to
	e.GET("/ws/:channelID", wsHandler(s))
	return s
}

func SetupHealthEndpoint(cfg config.Config, g *echo.Group, c *client.Client) {
	thttp.NewHealthHandler(service.NewHealthService(c, cfg.P4)).RegisterRoutes(g)
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
		return svr.Listen(ws, c.Param("channelID"))
	}
}

// PrintDev outputs some useful dev information such as http routes
// and current settings being used.
func PrintDev(e *echo.Echo) {
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
