package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/tonicpow/go-minercraft"
	gopaymail "github.com/tonicpow/go-paymail"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/data/mapi"
	"github.com/libsv/payd/data/paymail"
	paydSQL "github.com/libsv/payd/data/sqlite"
	"github.com/libsv/payd/data/sqlite/schema"
	"github.com/libsv/payd/service"
	"github.com/libsv/payd/service/ppctl"
	"github.com/libsv/payd/transports/http"

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

func main() {
	println("\033[32m" + banner + "\033[0m")
	cfg := config.NewViperConfig(appname).
		WithServer().
		WithDb().
		WithDeployment(appname).
		WithLog().
		WithPaymail().
		WithWallet().
		WithMapi()

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
	var paymentSender gopayd.PaymentSender
	var paymentOutputter gopayd.PaymentRequestOutputer
	if cfg.Paymail.UsePaymail {
		pCli, err := gopaymail.NewClient(nil, nil, nil)
		if err != nil {
			log.Fatalf("unable to create paymail client %s: ", err)
		}
		paymailStore := paymail.NewPaymail(cfg.Paymail, pCli)
		paymentSender = ppctl.NewPaymailPaymentService(paymailStore, cfg.Paymail)
		paymentOutputter = ppctl.NewPaymailOutputs(cfg.Paymail, paymailStore)
	} else {
		mapiCli, err := minercraft.NewClient(nil, nil, []*minercraft.Miner{
			{
				Name:  cfg.Mapi.MinerName,
				Token: cfg.Mapi.Token,
				URL:   cfg.Mapi.URL,
			},
		})
		if err != nil {
			log.Fatalf("error occurred: %s", err)
		}
		pkSvc := service.NewPrivateKeys(sqlLiteStore, cfg.Deployment.MainNet)
		mapiStore := mapi.NewBroadcast(cfg.Mapi, mapiCli)
		paymentSender = ppctl.NewPaymentMapiSender(mapiStore)
		paymentOutputter = ppctl.NewMapiOutputs(cfg.Server, pkSvc, &paydSQL.SQLiteTransacter{}, sqlLiteStore)
	}

	http.NewPaymentRequestHandler(
		ppctl.NewPaymentRequest(cfg.Wallet, cfg.Server, paymentOutputter, sqlLiteStore)).
		RegisterRoutes(g)
	http.NewPaymentHandler(
		ppctl.NewPayment(sqlLiteStore, sqlLiteStore, sqlLiteStore, paymentSender, &paydSQL.SQLiteTransacter{})).
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
