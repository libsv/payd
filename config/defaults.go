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

	// deployment
	viper.SetDefault(EnvEnvironment, "dev")
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

	// wallet
	viper.SetDefault(EnvNetwork, "regtest")
	viper.SetDefault(EnvWalletSpvRequired, true)
	viper.SetDefault(EnvPaymentExpiry, 24)
}
