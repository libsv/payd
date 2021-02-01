package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/libsv/go-payd/bip270"
	"github.com/libsv/go-payd/ipaymail"
	"github.com/spf13/viper"
	"github.com/tonicpow/go-paymail"
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
	v1.GET("/r/:paymentID", bip270.SolicitPaymentRequestHandler)
	v1.POST("/payment/:paymentID", bip270.PaymentHandler)
}

func main() {

	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Commit: %s\n", commit)

	// TODO: add paymail capabilities/etc. to key value db with expiry and remove from here
	// ------ get paymail capabilities (START) ------
	const handcashDomain = "handcash.io"
	const moneybuttonDomain = "moneybutton.com"

	// Load the client
	client, err := paymail.NewClient(nil, nil, nil)
	if err != nil {
		log.Fatalf("error loading client: %s", err.Error())
	}

	// Get the capabilities
	// This is required first to get the corresponding P2P PaymentResolution endpoint url

	// handcash // TODO: FIX not working 2021/01/24 17:13:11 error getting capabilities: invalid character '<' looking for beginning of value
	// var handcashCap *paymail.Capabilities
	// if handcashCap, err = client.GetCapabilities(handcashDomain, paymail.DefaultPort); err != nil {
	// 	log.Fatal("error getting capabilities: " + err.Error())
	// }
	// log.Println("found capabilities: ", len(handcashCap.Capabilities))

	// moneybutton
	var moneybuttonCap *paymail.Capabilities
	if moneybuttonCap, err = client.GetCapabilities(moneybuttonDomain, paymail.DefaultPort); err != nil {
		log.Fatal("error getting capabilities: " + err.Error())
	}
	log.Println("found capabilities: ", len(moneybuttonCap.Capabilities))

	// ipaymail.GlobalPaymailCapabilities[handcashDomain] = handcashCap
	ipaymail.GlobalPaymailCapabilities[moneybuttonDomain] = moneybuttonCap
	// ------ get paymail capabilities (END) ------

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
