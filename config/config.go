package config

import (
	"fmt"
	"regexp"
	"time"

	validator "github.com/theflyingcodr/govalidator"
)

// Environment variable constants.
const (
	EnvServerPort               = "server.port"
	EnvServerHost               = "server.host"
	EnvServerSwaggerEnabled     = "server.swagger.enabled"
	EnvServerSwaggerHost        = "server.swagger.host"
	EnvEnvironment              = "env.environment"
	EnvRegion                   = "env.region"
	EnvVersion                  = "env.version"
	EnvCommit                   = "env.commit"
	EnvBuildDate                = "env.builddate"
	EnvBitcoinNetwork           = "env.bitcoin.network"
	EnvLogLevel                 = "log.level"
	EnvDb                       = "db.type"
	EnvDbSchema                 = "db.schema.path"
	EnvDbDsn                    = "db.dsn"
	EnvDbMigrate                = "db.migrate"
	EnvHeadersClientAddress     = "headersclient.address"
	EnvHeadersClientTimeout     = "headersclient.timeout"
	EnvNetwork                  = "wallet.network"
	EnvWalletSpvRequired        = "wallet.spvrequired"
	EnvPaymentExpiry            = "wallet.paymentexpiry"
	EnvWalletPayoutLimitSats    = "wallet.payoutlimit.sats" // max allowed to be paid
	EnvWalletPayoutLimitEnabled = "wallet.payoutlimit.enabled"
	EnvDPPTimeout               = "dpp.timeout"
	EnvDPPHost                  = "dpp.host"
	EnvMAPIMinerName            = "mapi.minername"
	EnvMAPIURL                  = "mapi.minerurl"
	EnvMAPIToken                = "mapi.token"
	EnvMAPICallbackHost         = "mapi.callback.host"
	EnvSocketMaxMessageBytes    = "socket.maxmessage.bytes"
	EnvTransportHTTPEnabled     = "transport.http.enabled"
	EnvTransportSocketsEnabled  = "transport.sockets.enabled"
	EnvPeerChannelsHost         = "peerchannels.host"
	EnvPeerChannelsPath         = "peerchannels.path"
	EnvPeerChannelsTLS          = "peerchannels.tls"
	EnvPeerChannelsTTL          = "peerchannels.ttl.minutes"

	LogDebug = "debug"
	LogInfo  = "info"
	LogError = "error"
	LogWarn  = "warn"
)

// NetworkType is used to restrict the networks we can support.
type NetworkType string

// Supported bitcoin network types.
const (
	NetworkRegtest NetworkType = "regtest"
	NetworkSTN     NetworkType = "stn"
	NetworkTestnet NetworkType = "testnet"
	NetworkMainet  NetworkType = "mainnet"
)

func (n NetworkType) String() string {
	return string(n)
}

var reDbType = regexp.MustCompile(`sqlite|mysql|postgres`)

// DbType is used to restrict the dbs we can support.
type DbType string

// Supported database types.
const (
	DBSqlite   DbType = "sqlite"
	DBMySQL    DbType = "mysql"
	DBPostgres DbType = "postgres"
)

var reNetworks = regexp.MustCompile(`^(regtest|stn|testnet|mainnet)$`)

// Config returns strongly typed config values.
type Config struct {
	Logging       *Logging
	Server        *Server
	Deployment    *Deployment
	Db            *Db
	HeadersClient *HeadersClient
	Wallet        *Wallet
	PeerChannels  *PeerChannels
	DPP           *DPP
	Mapi          *MApi
	Socket        *Socket
	Transports    *Transports
}

// Validate will ensure the config matches certain parameters.
func (c *Config) Validate() error {
	vl := validator.New()
	if c.Db != nil {
		vl = vl.Validate("db.type", validator.MatchString(string(c.Db.Type), reDbType))
	}
	if c.Wallet != nil {
		vl = vl.Validate("wallet.network", validator.MatchString(string(c.Wallet.Network), reNetworks))
	}
	return vl.Err()
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

// Logging contains log configuration.
type Logging struct {
	Level string
}

// Server contains all settings required to run a web server.
type Server struct {
	Port     string
	Hostname string
	// SwaggerEnabled if true we will include an endpoint to serve swagger documents.
	SwaggerEnabled bool
	SwaggerHost    string
}

// Db contains database information.
type Db struct {
	Type       DbType
	SchemaPath string
	Dsn        string
	MigrateDb  bool
}

// HeadersClient contains HeadersClient information.
type HeadersClient struct {
	Address string
	Timeout int
}

// Wallet contains information relating to a payd installation.
type Wallet struct {
	Network             NetworkType
	SPVRequired         bool
	PaymentExpiryHours  int64
	PayoutLimitEnabled  bool
	PayoutLimitSatoshis uint64
}

// PeerChannels information relating to peer channel interactions.
type PeerChannels struct {
	// Host the peer channels host.
	Host string
	// Path to peer channels.
	Path string
	// TTL the life of the peer channel.
	TTL time.Duration
	// TLS if true will enable https / wss.
	TLS bool
}

// DPP contains information relating to a DPP interactions.
type DPP struct {
	Timeout    int
	ServerHost string
}

// MApi contains MAPI connection settings.
type MApi struct {
	MinerName    string
	URL          string
	Token        string
	CallbackHost string
}

// Socket contains the socket config for this server if running sockets.
type Socket struct {
	MaxMessageBytes  int
	ClientIdentifier string
}

// Transports enables or disables dpp transports.
type Transports struct {
	HTTPEnabled    bool
	SocketsEnabled bool
}

// ConfigurationLoader will load configuration items
// into a struct that contains a configuration.
type ConfigurationLoader interface {
	WithServer() ConfigurationLoader
	WithDb() ConfigurationLoader
	WithDeployment(app string) ConfigurationLoader
	WithLog() ConfigurationLoader
	WithWallet() ConfigurationLoader
	WithDPP() ConfigurationLoader
	WithHeadersClient() ConfigurationLoader
	WithSocket() ConfigurationLoader
	WithTransports() ConfigurationLoader
	WithMapi() ConfigurationLoader
	WithPeerChannels() ConfigurationLoader
	Load() *Config
}
