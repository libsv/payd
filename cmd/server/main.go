package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/libsv/payd/cmd/internal"
	"github.com/theflyingcodr/sockets/client"

	"github.com/libsv/payd/config"
	"github.com/libsv/payd/config/databases"
	_ "github.com/libsv/payd/docs"
	"github.com/libsv/payd/log"
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
		WithMapi().
		WithSocket().
		WithTransports().
		Load()
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

	e := internal.SetupEcho(log)

	if cfg.Server.SwaggerEnabled {
		internal.SetupSwagger(*cfg.Server, e)
	}
	// socket client server
	c := client.New(client.WithMaxMessageSize(10000), client.WithPongTimeout(360*time.Second))
	defer c.Close()

	g := e.Group("/")

	// setup transports
	internal.SetupHTTPEndpoints(*cfg, internal.SetupRestDeps(cfg, log, db, c), g)

	// setup sockets
	deps := internal.SetupSocketDeps(cfg, log, db, c)
	internal.SetupSocketClient(*cfg, deps, c)
	// setup socket endpoints
	internal.SetupSocketHTTPEndpoints(*cfg.Deployment, deps, g)

	if cfg.Deployment.IsDev() {
		internal.PrintDev(e)
	}
	go func() {
		log.Error(e.Start(cfg.Server.Port), "echo server failed")
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Error(err, "")
	}
}
