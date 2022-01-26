package config

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// ViperConfig contains viper based configuration data.
type ViperConfig struct {
	*Config
}

// NewViperConfig will setup and return a new viper based configuration handler.
func NewViperConfig(appname string) *ViperConfig {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return &ViperConfig{
		Config: &Config{},
	}
}

// WithServer will setup the web server configuration if required.
func (v *ViperConfig) WithServer() ConfigurationLoader {
	v.Server = &Server{
		Port:           viper.GetString(EnvServerPort),
		Hostname:       viper.GetString(EnvServerHost),
		SwaggerEnabled: viper.GetBool(EnvServerSwaggerEnabled),
		SwaggerHost:    viper.GetString(EnvServerSwaggerHost),
	}
	return v
}

// WithDeployment sets up the deployment configuration if required.
func (v *ViperConfig) WithDeployment(appName string) ConfigurationLoader {
	v.Deployment = &Deployment{
		Environment: viper.GetString(EnvEnvironment),
		Region:      viper.GetString(EnvRegion),
		Version:     viper.GetString(EnvVersion),
		Commit:      viper.GetString(EnvCommit),
		BuildDate:   viper.GetTime(EnvBuildDate),
		AppName:     appName,
	}
	return v
}

// WithLog sets up and returns log config.
func (v *ViperConfig) WithLog() ConfigurationLoader {
	v.Logging = &Logging{Level: viper.GetString(EnvLogLevel)}
	return v
}

// WithDb sets up and returns database configuration.
func (v *ViperConfig) WithDb() ConfigurationLoader {
	v.Db = &Db{
		Type:       DbType(viper.GetString(EnvDb)),
		Dsn:        viper.GetString(EnvDbDsn),
		SchemaPath: viper.GetString(EnvDbSchema),
		MigrateDb:  viper.GetBool(EnvDbMigrate),
	}
	return v
}

// WithHeadersClient sets up and returns headers client configuration.
func (v *ViperConfig) WithHeadersClient() ConfigurationLoader {
	v.HeadersClient = &HeadersClient{
		Address: viper.GetString(EnvHeadersClientAddress),
		Timeout: viper.GetInt(EnvHeadersClientTimeout),
	}
	return v
}

// WithWallet sets up and returns merchant wallet configuration.
func (v *ViperConfig) WithWallet() ConfigurationLoader {
	v.Wallet = &Wallet{
		Network:            NetworkType(viper.GetString(EnvNetwork)),
		SPVRequired:        viper.GetBool(EnvWalletSpvRequired),
		PaymentExpiryHours: viper.GetInt64(EnvPaymentExpiry),
	}
	return v
}

// WithP4 sets up and return p4 interface configuration.
func (v *ViperConfig) WithP4() ConfigurationLoader {
	v.P4 = &P4{
		ServerHost: viper.GetString(EnvP4Host),
		Timeout:    viper.GetInt(EnvP4Timeout),
	}
	return v
}

// WithMapi will setup Mapi settings.
func (v *ViperConfig) WithMapi() ConfigurationLoader {
	v.Mapi = &MApi{
		MinerName:    viper.GetString(EnvMAPIMinerName),
		URL:          viper.GetString(EnvMAPIURL),
		Token:        viper.GetString(EnvMAPIToken),
		CallbackHost: viper.GetString(EnvMAPICallbackHost),
	}
	return v
}

// WithSocket will setup Mapi settings.
func (v *ViperConfig) WithSocket() ConfigurationLoader {
	v.Socket = &Socket{ClientIdentifier: uuid.NewString()}
	return v
}

// WithTransports reads transport config.
func (v *ViperConfig) WithTransports() ConfigurationLoader {
	v.Transports = &Transports{
		HTTPEnabled:    viper.GetBool(EnvTransportHTTPEnabled),
		SocketsEnabled: viper.GetBool(EnvTransportSocketsEnabled),
	}
	return v
}

// WithPeerChannels reads peer channels config.
func (v *ViperConfig) WithPeerChannels() ConfigurationLoader {
	v.PeerChannels = &PeerChannels{
		Host: viper.GetString(EnvPeerChannelsHost),
		TTL:  time.Duration(viper.GetInt64(EnvPeerChannelsTTL)) * time.Minute,
	}
	return v
}

// Load will return the underlying config setup.
func (v *ViperConfig) Load() *Config {
	return v.Config
}
