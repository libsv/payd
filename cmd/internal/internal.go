package internal

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/InVisionApp/go-health/v2"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	paydSQL "github.com/libsv/payd/data/sqlite"
	"github.com/libsv/payd/docs"
	"github.com/libsv/payd/dpp"
	"github.com/libsv/payd/log"
	"github.com/libsv/payd/service"
	thttp "github.com/libsv/payd/transports/http"
	paydMiddleware "github.com/libsv/payd/transports/http/middleware"
	tsoc "github.com/libsv/payd/transports/sockets"
	socMiddleware "github.com/libsv/payd/transports/sockets/middleware"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/client"
	smw "github.com/theflyingcodr/sockets/middleware"
	"github.com/theflyingcodr/sockets/server"
)

// SetupEcho will set up and return an echo server.
func SetupEcho(cfg *config.Config, l log.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return cfg.Logging.Level != config.LogDebug
		},
	}))
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	p := prometheus.NewPrometheus("payd", nil)
	p.Use(e)
	e.HTTPErrorHandler = paydMiddleware.ErrorHandler(l)
	return e
}

// SetupSwagger will enable the swagger endpoints.
func SetupSwagger(cfg config.Server, e *echo.Echo) {
	docs.SwaggerInfo.Host = cfg.SwaggerHost
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}

// SetupHTTPEndpoints will register the http endpoints.
func SetupHTTPEndpoints(cfg config.Config, services *RestDeps, g *echo.Group) {
	// handlers
	thttp.NewInvoice(services.InvoiceService).
		RegisterRoutes(g)
	thttp.NewBalance(services.BalanceService).RegisterRoutes(g)
	thttp.NewProofs(services.ProofService).RegisterRoutes(g)
	thttp.NewPayments(services.PaymentService).RegisterRoutes(g)
	thttp.NewPaymentRequests(services.PaymentRequestService, cfg.DPP).RegisterRoutes(g)
	thttp.NewOwnersHandler(services.OwnerService).RegisterRoutes(g)
	thttp.NewUsersHandler(services.UserService).RegisterRoutes(g)
	thttp.NewPayHandler(services.PayService).RegisterRoutes(g)
	if cfg.Deployment.Environment == "local" {
		// ugly endpoint for regtest topup - local only!
		thttp.NewTransactions(services.TransactionService).RegisterRoutes(g)
	}
}

// SetupSocketClient will setup handlers and socket server.
func SetupSocketClient(cfg config.Config, deps *SocketDeps, c *client.Client) {
	lcfg := smw.NewLoggerConfig()
	lcfg.AddSkipper(func(msg *sockets.Message) bool {
		return cfg.Logging.Level != config.LogDebug
	})
	c.WithMiddleware(smw.PanicHandler,
		smw.Timeout(smw.NewTimeoutConfig()),
		smw.Metrics(),
		smw.Logger(lcfg),
		socMiddleware.IgnoreMyMessages(cfg.Socket),
		socMiddleware.WithAppIDPayD()).
		WithErrorHandler(socMiddleware.ErrorHandler).
		WithServerErrorHandler(socMiddleware.ErrorMsgHandler)

	// client handlers
	tsoc.NewPaymentRequest(&paydSQL.Transacter{}, deps.PaymentRequestService, deps.EnvelopeService, cfg.DPP).
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

// SetupHealthEndpoint setup the health check.
func SetupHealthEndpoint(cfg config.Config, g *echo.Group, c *client.Client, deps *SocketDeps) error {
	h := health.New()

	if err := dpp.NewHealthCheck(h, c, deps.InvoiceService, deps.ConnectService, cfg.DPP).Start(); err != nil {
		return errors.Wrap(err, "failed to start dpp health check")
	}

	thttp.NewHealthHandler(service.NewHealthService(h)).RegisterRoutes(g)

	return errors.Wrap(h.Start(), "failed to start health checker")
}

// ResumeActiveChannels resume listening to active peer channels.
func ResumeActiveChannels(deps *SocketDeps, l log.Logger) error {
	ctx := context.Background()
	channels, err := deps.PeerChannelsService.ActiveProofChannels(ctx)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		ch := channel
		if err := deps.PeerChannelsNotifyService.Subscribe(ctx, &ch); err != nil {
			l.Errorf(err, "failed to re-subscribe to channel %s", ch.ID)
		}
	}

	return nil
}

// ResumeSocketConnections resume socket connections with the DPP host.
func ResumeSocketConnections(deps *SocketDeps, cfg *config.DPP) error {
	u, err := url.Parse(cfg.ServerHost)
	if err != nil {
		return errors.Wrap(err, "failed to parse dpp host")
	}

	// No need to re-establish socket conn when running over http
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return nil
	}

	ctx := context.Background()
	invoices, err := deps.InvoiceService.InvoicesPending(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve invoices")
	}

	for _, invoice := range invoices {
		if time.Now().UTC().Unix() <= invoice.ExpiresAt.Time.UTC().Unix() {
			if err := deps.ConnectService.Connect(ctx, payd.ConnectArgs{
				InvoiceID: invoice.ID,
			}); err != nil {
				return errors.Wrapf(err, "failed to connect invoice %s", invoice.ID)
			}
		}
	}

	return nil
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
