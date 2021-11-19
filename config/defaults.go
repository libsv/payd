package config

import (
	"time"

	"github.com/spf13/viper"
)

// SetupDefaults will setup default config values.
// These can all be overwritten by environment variables.
func SetupDefaults() {
	// server
	viper.SetDefault(EnvServerPort, ":8443")
	viper.SetDefault(EnvServerHost, "payd:8443")
	viper.SetDefault(EnvServerSwaggerEnabled, true)
	viper.SetDefault(EnvServerSwaggerHost, "localhost:8443")

	// deployment
	viper.SetDefault(EnvEnvironment, "local")
	viper.SetDefault(EnvRegion, "test")
	viper.SetDefault(EnvCommit, "test")
	viper.SetDefault(EnvVersion, "test")
	viper.SetDefault(EnvBuildDate, time.Now().UTC())

	// logging
	viper.SetDefault(EnvLogLevel, "info")

	// db
	viper.SetDefault(EnvDb, "sqlite")
	viper.SetDefault(EnvDbDsn, "file:data/wallet.db?_foreign_keys=true&pooled=true")
	viper.SetDefault(EnvDbSchema, "data/sqlite/migrations")
	viper.SetDefault(EnvDbMigrate, true)

	// headers client
	viper.SetDefault(EnvHeadersClientAddress, "http://headersv:8080")
	viper.SetDefault(EnvHeadersClientTimeout, 30)

	// p4
	viper.SetDefault(EnvP4Timeout, 30)
	viper.SetDefault(EnvP4Host, "p4:8442")

	// wallet
	viper.SetDefault(EnvNetwork, string(NetworkRegtest))
	viper.SetDefault(EnvWalletSpvRequired, false)
	viper.SetDefault(EnvPaymentExpiry, 24)

	// mapi
	viper.SetDefault(EnvMAPIMinerName, "local-mapi")
	viper.SetDefault(EnvMAPIURL, "http://mapi:80")
	viper.SetDefault(EnvMAPIToken, "")

	// Socket settings
	viper.SetDefault(EnvSocketMaxMessageBytes, 10000)

	// Transport settings
	viper.SetDefault(EnvTransportHTTPEnabled, true)
	viper.SetDefault(EnvTransportSocketsEnabled, true)
}
