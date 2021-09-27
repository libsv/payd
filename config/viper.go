package config

import (
	"strings"

	"github.com/spf13/viper"
)

// NewViperConfig will setup and return a new viper based configuration handler.
func NewViperConfig(appname string) *Config {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return &Config{}
}

// WithServer will setup the web server configuration if required.
func (c *Config) WithServer() *Config {
	c.Server = &Server{
		Port:     viper.GetString(EnvServerPort),
		Hostname: viper.GetString(EnvServerHost),
	}
	return c
}

// WithDeployment sets up the deployment configuration if required.
func (c *Config) WithDeployment(appName string) *Config {
	c.Deployment = &Deployment{
		Environment: viper.GetString(EnvEnvironment),
		Region:      viper.GetString(EnvRegion),
		Version:     viper.GetString(EnvVersion),
		Commit:      viper.GetString(EnvCommit),
		BuildDate:   viper.GetTime(EnvBuildDate),
		AppName:     appName,
		MainNet:     viper.GetBool(EnvMainNet),
	}
	return c
}

// WithLog sets up and returns log config.
func (c *Config) WithLog() *Config {
	c.Logging = &Logging{Level: viper.GetString(EnvLogLevel)}
	return c
}

// WithDb sets up and returns database configuration.
func (c *Config) WithDb() *Config {
	c.Db = &Db{
		Type:       DbType(viper.GetString(EnvDb)),
		Dsn:        viper.GetString(EnvDbDsn),
		SchemaPath: viper.GetString(EnvDbSchema),
		MigrateDb:  viper.GetBool(EnvDbMigrate),
	}
	return c
}

// WithHeadersClient sets up and returns headers client configuration.
func (c *Config) WithHeadersClient() *Config {
	c.HeadersClient = &HeadersClient{
		Address: viper.GetString(EnvHeadersClientAddress),
		Timeout: viper.GetInt(EnvHeadersClientTimeout),
	}
	return c
}

// WithWallet sets up and returns merchant wallet configuration.
func (c *Config) WithWallet() *Config {

	c.Wallet = &Wallet{
		Network:            viper.GetString(EnvNetwork),
		SPVRequired:        viper.GetBool(EnvWalletSpvRequired),
		PaymentExpiryHours: viper.GetInt64(EnvPaymentExpiry),
	}
	return c
}

// WithP4 sets up and return p4 interface configuration.
func (c *Config) WithP4() *Config {
	c.P4 = &P4{
		Timeout: viper.GetInt(EnvP4Timeout),
	}
	return c
}

// WithMapi will setup Mapi settings.
func (c *Config) WithMapi() *Config {
	viper.SetDefault(EnvMAPIMinerName, "local-mapi")
	viper.SetDefault(EnvMAPIURL, "http://mapi:80")
	viper.SetDefault(EnvMAPIToken, "")
	c.Mapi = &MApi{
		MinerName: viper.GetString(EnvMAPIMinerName),
		URL:       viper.GetString(EnvMAPIURL),
		Token:     viper.GetString(EnvMAPIToken),
	}
	return c
}
