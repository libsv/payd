package minercraft

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gojektech/heimdall/v6"
	"github.com/gojektech/heimdall/v6/httpclient"
)

// HTTPInterface is used for the http client (mocking heimdall)
type HTTPInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the parent struct that contains the miner clients and list of miners to use
type Client struct {
	httpClient HTTPInterface  // Interface for all HTTP requests
	miners     []*Miner       // List of loaded miners
	Options    *ClientOptions // Client options config
}

// AddMiner will add a new miner to the list of miners
func (c *Client) AddMiner(miner Miner) error {

	// Make sure we have the basic requirements
	if len(miner.Name) == 0 {
		return errors.New("missing miner name")
	} else if len(miner.URL) == 0 {
		return errors.New("missing miner url")
	}

	// Check if a miner with that name already exists
	existingMiner := c.MinerByName(miner.Name)
	if existingMiner != nil {
		return fmt.Errorf("miner %s already exists", miner.Name)
	}

	// Check if a miner with the minerID already exists
	if len(miner.MinerID) > 0 {
		if existingMiner = c.MinerByID(miner.MinerID); existingMiner != nil {
			return fmt.Errorf("miner %s already exists", miner.MinerID)
		}
	}

	// Ensure that we have a protocol
	if !strings.Contains(miner.URL, "http") {
		return fmt.Errorf("miner %s is missing http from url", miner.Name)
	}

	// Test parsing the url
	parsedURL, err := url.Parse(miner.URL)
	if err != nil {
		return err
	}
	miner.URL = parsedURL.String()

	// Append the new miner
	c.miners = append(c.miners, &miner)
	return nil
}

// RemoveMiner will remove a miner from the list
func (c *Client) RemoveMiner(miner *Miner) bool {
	for i, m := range c.miners {
		if m.Name == miner.Name || m.MinerID == miner.MinerID {
			c.miners[i] = c.miners[len(c.miners)-1]
			c.miners = c.miners[:len(c.miners)-1]
			return true
		}
	}
	// Miner not found
	return false
}

// MinerByName will return a miner given a name
func (c *Client) MinerByName(name string) *Miner {
	return MinerByName(c.miners, name)
}

// MinerByName will return a miner from a given set of miners
func MinerByName(miners []*Miner, minerName string) *Miner {
	for index, miner := range miners {
		if strings.EqualFold(minerName, miner.Name) {
			return miners[index]
		}
	}
	return nil
}

// MinerByID will return a miner given a miner id
func (c *Client) MinerByID(minerID string) *Miner {
	return MinerByID(c.miners, minerID)
}

// MinerByID will return a miner from a given set of miners
func MinerByID(miners []*Miner, minerID string) *Miner {
	for index, miner := range miners {
		if strings.EqualFold(minerID, miner.MinerID) {
			return miners[index]
		}
	}
	return nil
}

// MinerUpdateToken will find a miner by name and update the token
func (c *Client) MinerUpdateToken(name, token string) {
	if miner := c.MinerByName(name); miner != nil {
		miner.Token = token
	}
}

// Miners will return the list of miners
func (c *Client) Miners() []*Miner {
	return c.miners
}

// UserAgent will return the user agent
func (c *Client) UserAgent() string {
	return c.Options.UserAgent
}

// ClientOptions holds all the configuration for connection, dialer and transport
type ClientOptions struct {
	BackOffExponentFactor          float64       `json:"back_off_exponent_factor"`
	BackOffInitialTimeout          time.Duration `json:"back_off_initial_timeout"`
	BackOffMaximumJitterInterval   time.Duration `json:"back_off_maximum_jitter_interval"`
	BackOffMaxTimeout              time.Duration `json:"back_off_max_timeout"`
	DialerKeepAlive                time.Duration `json:"dialer_keep_alive"`
	DialerTimeout                  time.Duration `json:"dialer_timeout"`
	RequestRetryCount              int           `json:"request_retry_count"`
	RequestTimeout                 time.Duration `json:"request_timeout"`
	TransportExpectContinueTimeout time.Duration `json:"transport_expect_continue_timeout"`
	TransportIdleTimeout           time.Duration `json:"transport_idle_timeout"`
	TransportMaxIdleConnections    int           `json:"transport_max_idle_connections"`
	TransportTLSHandshakeTimeout   time.Duration `json:"transport_tls_handshake_timeout"`
	UserAgent                      string        `json:"user_agent"`
}

// DefaultClientOptions will return a ClientOptions struct with the default settings.
// Useful for starting with the default and then modifying as needed
func DefaultClientOptions() (clientOptions *ClientOptions) {
	return &ClientOptions{
		BackOffExponentFactor:          2.0,
		BackOffInitialTimeout:          2 * time.Millisecond,
		BackOffMaximumJitterInterval:   2 * time.Millisecond,
		BackOffMaxTimeout:              10 * time.Millisecond,
		DialerKeepAlive:                20 * time.Second,
		DialerTimeout:                  5 * time.Second,
		RequestRetryCount:              2,
		RequestTimeout:                 30 * time.Second,
		TransportExpectContinueTimeout: 3 * time.Second,
		TransportIdleTimeout:           20 * time.Second,
		TransportMaxIdleConnections:    10,
		TransportTLSHandshakeTimeout:   5 * time.Second,
		UserAgent:                      defaultUserAgent,
	}
}

// NewClient creates a new client for requests
//
// clientOptions: inject custom client options on load
// customHTTPClient: use your own custom HTTP client
// customMiners: use your own custom list of miners
func NewClient(clientOptions *ClientOptions, customHTTPClient HTTPInterface,
	customMiners []*Miner) (client ClientInterface, err error) {

	// Create the new client
	return createClient(clientOptions, customHTTPClient, customMiners)
}

// createClient will make a new http client based on the options provided
func createClient(options *ClientOptions, customHTTPClient HTTPInterface, customMiners []*Miner) (c *Client, err error) {

	// Create a client
	c = new(Client)

	// Set options (either default or user modified)
	if options == nil {
		options = DefaultClientOptions()
	}

	// Set the options
	c.Options = options

	// Load custom vs pre-defined
	if len(customMiners) > 0 {
		c.miners = customMiners
	} else {
		if c.miners, err = DefaultMiners(); err != nil {
			return nil, err
		}
	}

	// Is there a custom HTTP client to use?
	if customHTTPClient != nil {
		c.httpClient = customHTTPClient
		return
	}

	// dial is the net dialer for clientDefaultTransport
	dial := &net.Dialer{KeepAlive: options.DialerKeepAlive, Timeout: options.DialerTimeout}

	// clientDefaultTransport is the default transport struct for the HTTP client
	clientDefaultTransport := &http.Transport{
		DialContext:           dial.DialContext,
		ExpectContinueTimeout: options.TransportExpectContinueTimeout,
		IdleConnTimeout:       options.TransportIdleTimeout,
		MaxIdleConns:          options.TransportMaxIdleConnections,
		Proxy:                 http.ProxyFromEnvironment,
		TLSHandshakeTimeout:   options.TransportTLSHandshakeTimeout,
	}

	// Determine the strategy for the http client
	if options.RequestRetryCount <= 0 {

		// no retry enabled
		c.httpClient = httpclient.NewClient(
			httpclient.WithHTTPTimeout(options.RequestTimeout),
			httpclient.WithHTTPClient(&http.Client{
				Transport: clientDefaultTransport,
				Timeout:   options.RequestTimeout,
			}),
		)
		return
	}

	// Retry enabled - create exponential back-off
	c.httpClient = httpclient.NewClient(
		httpclient.WithHTTPTimeout(options.RequestTimeout),
		httpclient.WithRetrier(heimdall.NewRetrier(
			heimdall.NewExponentialBackoff(
				options.BackOffInitialTimeout,
				options.BackOffMaxTimeout,
				options.BackOffExponentFactor,
				options.BackOffMaximumJitterInterval,
			))),
		httpclient.WithRetryCount(options.RequestRetryCount),
		httpclient.WithHTTPClient(&http.Client{
			Transport: clientDefaultTransport,
			Timeout:   options.RequestTimeout,
		}),
	)

	return
}

// DefaultMiners will parse the config JSON and return a list of miners
func DefaultMiners() (miners []*Miner, err error) {
	err = json.Unmarshal([]byte(KnownMiners), &miners)
	return
}
