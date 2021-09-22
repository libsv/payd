package config

import (
	"fmt"
	"net/url"
	"path"

	"github.com/spf13/viper"
)

type ContextKey string

const (
	CfgCurrentContext            = "current.context"
	CfgWalletHost     ContextKey = "wallet.host"
	CfgWalletPort     ContextKey = "wallet.port"
	CfgAccountName    ContextKey = "account.name"
)

func (c ContextKey) Key() string {
	return fmt.Sprintf("contexts.%s.%s", viper.GetString(CfgCurrentContext), c)
}

type Config struct {
	*Context
	Contexts []*Context
}

type Context struct {
	Account *Account
	Wallet  *Wallet
}

func NewConfig() *Config {
	return &Config{
		Context:  &Context{},
		Contexts: make([]*Context, 0),
	}
}

type Wallet struct {
	host string
	port string
}

func (c *Config) WithWallet() *Config {
	viper.SetDefault(CfgWalletHost.Key(), "http://payd:8443")
	viper.SetDefault(CfgWalletPort.Key(), ":8443")
	c.Wallet = &Wallet{
		host: viper.GetString(CfgWalletHost.Key()),
		port: viper.GetString(CfgWalletPort.Key()),
	}
	return c
}

func (w *Wallet) URLFor(parts ...string) string {
	url, err := url.Parse(w.host)
	if err != nil {
		panic(err)
	}

	url.Path = path.Join(parts...)

	return url.String()
}

type Account struct {
	Name string
}

func (c *Config) WithAccount() *Config {
	viper.SetDefault(CfgAccountName.Key(), "me")
	c.Account = &Account{
		Name: viper.GetString(CfgAccountName.Key()),
	}

	return c
}
