package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/libsv/go-payd/ipaymail"
	"github.com/libsv/go-payd/ppctl"
	"github.com/spf13/viper"
)

const appname = "payd"

// Version & commit strings injected at build with -ldflags -X...
var version string
var commit string

func v1Routes(e *echo.Echo) {
	apiVersion := "/v1"

	v1 := e.Group(apiVersion)

	v1.POST("/invoice", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!") // TODO: implement
	})
	v1.GET("/r/:paymentID", ppctl.SolicitPaymentRequestHandler)
	v1.POST("/payment/:paymentID", ppctl.PaymentHandler)
}

func main() {

	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Commit: %s\n", commit)

	viper.SetDefault("bsvalias", "1.0")

	ipaymail.PaymailInit()
	// TODO: add paymail capabilities/etc. to key value db with expiry and remove from here (REDIS)
	const handcashDomain = "handcash.io"
	const moneybuttonDomain = "moneybutton.com"
	var err error

	ipaymail.GlobalPaymailCapabilities[handcashDomain], err = ipaymail.GetCapabilities(handcashDomain, false)
	if err != nil {
		log.Fatal("error getting capabilities: " + err.Error())
	}
	ipaymail.GlobalPaymailCapabilities[moneybuttonDomain], err = ipaymail.GetCapabilities(moneybuttonDomain, false)
	if err != nil {
		log.Fatal("error getting capabilities: " + err.Error())
	}

	// viperSetup() // TODO: setup viper prooperly

	//Echo library managed http routes
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "payd")
	})

	//Load the V1 routes
	v1Routes(e)

	//Start the server
	e.Logger.Fatal(e.Start(":1323"))
}

func viperSetup() {
	viper.SetConfigName("config")
	viper.SetConfigType("ini")
	viper.AddConfigPath(fmt.Sprintf("/etc/%s/", appname))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", appname))
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			viper.Set("debug", false)
			viper.Set("port", 1323)
		} else {
			panic(fmt.Errorf("Fatal error config file: %s", err))
		}
	}

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()
}
