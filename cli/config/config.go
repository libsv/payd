package config

import (
	"fmt"
	"net/url"
	"path"

	"github.com/spf13/viper"
)

type ContextKey string

const (
	CfgContexts                  = "contexts"
	CfgCurrentContext            = "current.context"
	CfgWalletHost     ContextKey = "wallet.host"
	CfgWalletPort     ContextKey = "wallet.port"
	CfgAccountName    ContextKey = "account.name"
)

func (c ContextKey) Key() string {
	return fmt.Sprintf("contexts.%s.%s", viper.GetString(CfgCurrentContext), c)
}

func (c ContextKey) KeyFor(name string) string {
	return fmt.Sprintf("contexts.%s.%s", name, c)
}

type Config struct {
	CurrentContext string
	*Context
	Contexts map[string]*Context
}

type Context struct {
	Account *Account
	Wallet  *Wallet
}

func NewConfig() *Config {
	return &Config{
		CurrentContext: viper.GetString(CfgCurrentContext),
		Context:        &Context{},
		Contexts:       make(map[string]*Context),
	}
}

type Wallet struct {
	host string
	port string
}

func (c *Config) WithWallet() *Config {
	viper.SetDefault(CfgWalletHost.KeyFor(c.CurrentContext), "http://payd:8443")
	viper.SetDefault(CfgWalletPort.KeyFor(c.CurrentContext), ":8443")
	c.Wallet = &Wallet{
		host: viper.GetString(CfgWalletHost.KeyFor(c.CurrentContext)),
		port: viper.GetString(CfgWalletPort.KeyFor(c.CurrentContext)),
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
	viper.SetDefault(CfgAccountName.KeyFor(c.CurrentContext), "me")
	c.Account = &Account{
		Name: viper.GetString(CfgAccountName.KeyFor(c.CurrentContext)),
	}

	return c
}

func (c *Config) WithContexts() *Config {
	mm := viper.Get(CfgContexts).(map[string]interface{})
	for k := range mm {
		cfg := NewConfig()
		cfg.CurrentContext = k

		cfg.WithWallet().WithAccount()

		c.Contexts[k] = cfg.Context
	}

	return c
}

func (c *Config) HasContext(name string) bool {
	_, ok := c.Contexts[name]
	return ok
}

func (c *Config) LoadContext(name string) bool {
	if !c.HasContext(name) {
		return false
	}

	c.CurrentContext = name
	c.Context = c.Contexts[name]
	return true
}
