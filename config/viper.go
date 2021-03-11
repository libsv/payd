package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func NewViperConfig(appname string) *Config {
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
			log.Fatalf("Fatal error config file: %s", err)
		}
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return &Config{}
}

// WithServer will setup the web server configuration if required.
func (c *Config) WithServer() *Config {
	viper.SetDefault(EnvServerPort, ":8442")
	viper.SetDefault(EnvServerHost, "localhost:8442")
	c.Server = &Server{
		Port:     viper.GetString(EnvServerPort),
		Hostname: viper.GetString(EnvServerHost),
	}
	return c
}

// WithDeployment sets up the deployment configuration if required.
func (c *Config) WithDeployment(appName string) *Config {
	viper.SetDefault(EnvEnvironment, "dev")
	viper.SetDefault(EnvRegion, "test")
	viper.SetDefault(EnvCommit, "test")
	viper.SetDefault(EnvVersion, "test")
	viper.SetDefault(EnvBuildDate, time.Now().UTC())
	viper.SetDefault(EnvMainNet, false)

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

func (c *Config) WithLog() *Config {
	viper.SetDefault(EnvLogLevel, "info")
	c.Logging = &Logging{Level: viper.GetString(EnvLogLevel)}
	return c
}

// WithDb sets up and returns database configuration.
func (c *Config) WithDb() *Config {
	viper.SetDefault(EnvDb, "sqlite")
	viper.SetDefault("db.dsn", "file:data/wallet.db?cache=shared&_foreign_keys=true;")
	c.Db = &Db{
		Type: viper.GetString(EnvDb),
		Dsn:  viper.GetString(EnvDbDsn),
	}
	return c
}

// WithPaymail sets up and returns paymail configuration.
func (c *Config) WithPaymail() *Config {
	viper.SetDefault(EnvPaymailEnabled, false)
	viper.SetDefault(EnvPaymailAddress, "test@test.com")
	viper.SetDefault(EnvPaymailIsBeta, false)
	c.Paymail = &Paymail{
		UsePaymail: viper.GetBool(EnvPaymailEnabled),
		IsBeta:     viper.GetBool(EnvPaymailIsBeta),
		Address:    viper.GetString(EnvPaymailAddress),
	}
	return c
}

// WithWallet sets up and returns merchant wallet configuration.
func (c *Config) WithWallet() *Config {
	viper.SetDefault(EnvNetwork, "bitcoin-sv")
	viper.SetDefault(EnvMerchantName, "payd")
	viper.SetDefault(EnvAvatarURL, "https://bit.ly/3c4iaup")
	c.Wallet = &Wallet{
		Network:           viper.GetString(EnvNetwork),
		MerchantAvatarURL: viper.GetString(EnvAvatarURL),
		MerchantName:      viper.GetString(EnvMerchantName),
	}
	return c
}

// WithMapi will setup Mapi settings.
func (c *Config) WithMapi() *Config {
	viper.SetDefault(EnvMAPIMinerName, "local-mapi")
	viper.SetDefault(EnvMAPIURL, "mapi:9004")
	viper.SetDefault(EnvMAPIToken, "")
	c.Mapi = &MApi{
		MinerName: viper.GetString(EnvMAPIMinerName),
		URL:       viper.GetString(EnvMAPIURL),
		Token:     viper.GetString(EnvMAPIToken),
	}
	return c
}
