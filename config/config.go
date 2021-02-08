package config

import (
	"fmt"
	"time"
)

const (
	EnvServerPort  = "server.port"
	EnvServerHost  = "server.host"
	EnvEnvironment = "env.environment"
	EnvRegion      = "env.region"
	EnvVersion     = "env.version"
	EnvCommit      = "env.commit"
	EnvBuildDate   = "env.builddate"
	EnvLogLevel    = "log.level"
	EnvDb          = "db.type"
	EnvDbDsn       = "db.dsn"
	EnvUsePaymail  = "paymail.enabled"

	LogDebug = "debug"
	LogInfo  = "info"
	LogError = "error"
	LogWarn  = "warn"
)

// Config returns strongly typed config values.
type Config struct {
	Logging    *Logging
	Server     *Server
	Deployment *Deployment
	Db         *Db
	Paymail    *Paymail
}

// Deployment contains information relating to the current
// deployed instance.
type Deployment struct {
	Environment string
	AppName     string
	Region      string
	Version     string
	Commit      string
	BuildDate   time.Time
}

// IsDev determines if this app is running on a dev environment.
func (d *Deployment) IsDev() bool {
	return d.Environment == "dev"
}

func (d *Deployment) String() string {
	return fmt.Sprintf("Environment: %s \n AppName: %s\n Region: %s\n Version: %s\n Commit:%s\n BuildDate: %s\n",
		d.Environment, d.AppName, d.Region, d.Version, d.Commit, d.BuildDate)
}

type Logging struct {
	Level string
}

// Server contains all settings required to run a web server.
type Server struct {
	Port     string
	Hostname string
}

// Db contains database information.
type Db struct {
	Type string
	Dsn  string
}

// Paymail settings relating to paymail.
type Paymail struct {
	UsePaymail bool
}

// ConfigurationLoader will load configuration items
// into a struct that contains a configuration.
type ConfigurationLoader interface {
	WithServer() *Config
	WithDb() *Config
	WithDeployment(app string) *Config
	WithLog() *Config
	WithPaymail() *Config
}