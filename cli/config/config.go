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
	CfgPaydHost       ContextKey = "payd.host"
	CfgPaydPort       ContextKey = "payd.port"
	CfgP4Host         ContextKey = "p4.host"
	CfgP4Port         ContextKey = "p4.port"
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
	Payd    *Payd
	P4      *P4
}

func NewConfig() *Config {
	viper.SetDefault(CfgCurrentContext, "me")
	return &Config{
		CurrentContext: viper.GetString(CfgCurrentContext),
		Context:        &Context{},
		Contexts:       make(map[string]*Context),
	}
}

type Payd struct {
	host string
	port string
}

func (c *Config) WithPayd() *Config {
	viper.SetDefault(CfgPaydHost.KeyFor(c.CurrentContext), "http://payd:8443")
	viper.SetDefault(CfgPaydPort.KeyFor(c.CurrentContext), ":8443")
	c.Payd = &Payd{
		host: viper.GetString(CfgPaydHost.KeyFor(c.CurrentContext)),
		port: viper.GetString(CfgPaydPort.KeyFor(c.CurrentContext)),
	}
	return c
}

type P4 struct {
	host string
	port string
}

func (c *Config) WithP4() *Config {
	viper.SetDefault(CfgP4Host.KeyFor(c.CurrentContext), "http://p4:8445")
	viper.SetDefault(CfgP4Port.KeyFor(c.CurrentContext), "://p4:8445")
	c.P4 = &P4{
		host: viper.GetString(CfgP4Host.KeyFor(c.CurrentContext)),
		port: viper.GetString(CfgP4Port.KeyFor(c.CurrentContext)),
	}

	return c
}

func (w *Payd) URLFor(parts ...string) string {
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

		cfg.WithPayd().WithAccount()

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
