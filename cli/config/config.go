package config

import (
	"fmt"
	"net/url"
	"path"

	"github.com/spf13/viper"
)

// ContextKey tpye context key.
type ContextKey string

// Configuration keys.
const (
	CfgContexts                  = "contexts"
	CfgCurrentContext            = "current.context"
	CfgPaydHost       ContextKey = "payd.host"
	CfgPaydPort       ContextKey = "payd.port"
	CfgP4Host         ContextKey = "p4.host"
	CfgP4Port         ContextKey = "p4.port"
	CfgAccountName    ContextKey = "account.name"
)

// Key returns the viper key for the current context.
func (c ContextKey) Key() string {
	return fmt.Sprintf("contexts.%s.%s", viper.GetString(CfgCurrentContext), c)
}

// KeyFor returns the context aware viper key for a provided context.
func (c ContextKey) KeyFor(name string) string {
	return fmt.Sprintf("contexts.%s.%s", name, c)
}

// Config holds config information.
type Config struct {
	CurrentContext string
	*Context
	Contexts map[string]*Context
}

// Context holds contextual configuration.
type Context struct {
	Account *Account
	Payd    *Payd
	P4      *P4
}

// NewConfig builds a new config.
func NewConfig() *Config {
	viper.SetDefault(CfgCurrentContext, "me")
	return &Config{
		CurrentContext: viper.GetString(CfgCurrentContext),
		Context:        &Context{},
		Contexts:       make(map[string]*Context),
	}
}

// Payd holds payd settings.
type Payd struct {
	host string
	port string
}

// WithPayd loads the current context's payd settings.
func (c *Config) WithPayd() *Config {
	viper.SetDefault(CfgPaydHost.KeyFor(c.CurrentContext), "http://payd:8443")
	viper.SetDefault(CfgPaydPort.KeyFor(c.CurrentContext), ":8443")
	c.Payd = &Payd{
		host: viper.GetString(CfgPaydHost.KeyFor(c.CurrentContext)),
		port: viper.GetString(CfgPaydPort.KeyFor(c.CurrentContext)),
	}
	return c
}

// URLFor returns a url for the parts provided, from the configurations host settings.
func (p *Payd) URLFor(parts ...string) string {
	url, err := url.Parse(p.host)
	if err != nil {
		panic(err)
	}

	url.Path = path.Join(parts...)

	return url.String()
}

// P4 holds p4 settings.
type P4 struct {
	host string
	port string
}

// WithP4 loads the current context's p4 settings.
func (c *Config) WithP4() *Config {
	viper.SetDefault(CfgP4Host.KeyFor(c.CurrentContext), "http://p4:8445")
	viper.SetDefault(CfgP4Port.KeyFor(c.CurrentContext), "://p4:8445")
	c.P4 = &P4{
		host: viper.GetString(CfgP4Host.KeyFor(c.CurrentContext)),
		port: viper.GetString(CfgP4Port.KeyFor(c.CurrentContext)),
	}

	return c
}

// URLFor returns a url for the parts provided, from the configurations host settings.
func (p *P4) URLFor(parts ...string) string {
	url, err := url.Parse(p.host)
	if err != nil {
		panic(err)
	}

	url.Path = path.Join(parts...)

	return url.String()
}

// Account holds account settings.
type Account struct {
	Name string
}

// WithAccount loads the current context's account settings.
func (c *Config) WithAccount() *Config {
	viper.SetDefault(CfgAccountName.KeyFor(c.CurrentContext), "me")
	c.Account = &Account{
		Name: viper.GetString(CfgAccountName.KeyFor(c.CurrentContext)),
	}

	return c
}

// WithContexts loads all contexts into the config.
func (c *Config) WithContexts() *Config {
	mm := viper.Get(CfgContexts).(map[string]interface{})
	for k := range mm {
		cfg := NewConfig()
		cfg.CurrentContext = k

		cfg.WithPayd().WithP4().WithAccount()

		c.Contexts[k] = cfg.Context
	}

	return c
}

// HasContext returns a boolean indiciating if the provided context exists.
func (c *Config) HasContext(name string) bool {
	_, ok := c.Contexts[name]
	return ok
}

// LoadContext loads and applies a context for the config.
func (c *Config) LoadContext(name string) bool {
	if !c.HasContext(name) {
		return false
	}

	c.CurrentContext = name
	c.Context = c.Contexts[name]
	return true
}
