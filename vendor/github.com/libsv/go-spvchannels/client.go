// ISC License
//
// Copyright (c) 2018-2020 The libsv developers
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// Package spvchannels is an golang implementation of the spv channel client.
//
// It implement all the rest api endpoints and the weboscket client to
// listen to notifications from channel in real time.
//
// Using the combination of the notification websocket and the rest api
// to pull unread message, users can have a real time message channel.
//
// SPV Channel BRFC
//
// https://github.com/bitcoin-sv-specs/brfc-spvchannels
//
// SPV Channel server
//
// https://github.com/bitcoin-sv/spvchannels-reference
package spvchannels

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"
)

// NotificationHandlerFunc is a callback to process websocket messages
//    ctx : the handling context
//    t   : message type
//    msg : message content
//    err : message error
type NotificationHandlerFunc = func(ctx context.Context, t int, msg []byte, err error) error

// ErrWSClose can be returned by a NotificationHandlerFunc to instruct the
// socket client that we are finished processing messages and to close.
//
// This could be emitted as a result of a server sending a message with a payload
// of 'close stream' or equivalent.
type ErrWSClose struct {
	error
}

// ErrorHandlerFunc is a callback to handle the error after processing the message
//    err : the error to handle
type ErrorHandlerFunc func(err error)

// spvConfig hold configuration for rest api connection
type spvConfig struct {
	insecure   bool // equivalent curl -k
	tls        bool
	baseURL    string
	path       string
	version    string
	user       string
	passwd     string
	token      string
	channelID  string
	procces    NotificationHandlerFunc
	errHandler ErrorHandlerFunc
}

func (s spvConfig) httpScheme() string {
	if s.tls {
		return "https"
	}
	return "http"
}

func (s spvConfig) wsScheme() string {
	if s.tls {
		return "wss"
	}
	return "ws"
}

// SPVConfigFunc set the rest api configuration
type SPVConfigFunc func(c *spvConfig)

// WithInsecure skip the TLS check (for dev only)
func WithInsecure() SPVConfigFunc {
	return func(c *spvConfig) {
		c.insecure = true
	}
}

// WithNoTLS use http and ws in place of https and wss (for dev only)
func WithNoTLS() SPVConfigFunc {
	return func(c *spvConfig) {
		c.tls = false
	}
}

// WithBaseURL provide base url (domain:port) for the rest api
func WithBaseURL(url string) SPVConfigFunc {
	return func(c *spvConfig) {
		c.baseURL = url
	}
}

// WithPath provide a path on the hosting service (/peerchannels)
func WithPath(path string) SPVConfigFunc {
	return func(c *spvConfig) {
		c.path = path
	}
}

// WithVersion provide version string for the rest api
func WithVersion(v string) SPVConfigFunc {
	return func(c *spvConfig) {
		c.version = v
	}
}

// WithUser provide username for rest basic authentification
func WithUser(userName string) SPVConfigFunc {
	return func(c *spvConfig) {
		c.user = userName
	}
}

// WithPassword provide password for rest basic authentification
func WithPassword(p string) SPVConfigFunc {
	return func(c *spvConfig) {
		c.passwd = p
	}
}

// WithToken provide token for rest token bearer authentification
func WithToken(t string) SPVConfigFunc {
	return func(c *spvConfig) {
		c.token = t
	}
}

// WithChannelID provide channel id for websocket notification
func WithChannelID(id string) SPVConfigFunc {
	return func(c *spvConfig) {
		c.channelID = id
	}
}

// WithWebsocketCallBack provide the callback function to process notification messages
func WithWebsocketCallBack(f NotificationHandlerFunc) SPVConfigFunc {
	return func(c *spvConfig) {
		c.procces = f
	}
}

// WithErrorHandler can be provided with a function used to handle
// errors when processing Socket message callbacks.
//
// Here you could log the errors, send to another system to drop them etc.
func WithErrorHandler(e ErrorHandlerFunc) SPVConfigFunc {
	return func(c *spvConfig) {
		c.errHandler = e
	}
}

func defaultSPVConfig() *spvConfig {
	// Set the default options
	cfg := &spvConfig{
		insecure:  false,
		tls:       true,
		baseURL:   "localhost:5010",
		version:   "v1",
		user:      "dev",
		passwd:    "dev",
		token:     "",
		channelID: "",
		procces: func(ctx context.Context, t int, msg []byte, err error) error {
			return err
		},
		errHandler: func(err error) {
			fmt.Printf("received err: %s\n", err)
		},
	}
	return cfg
}

// Client hold rest api configuration and http connection
type Client struct {
	cfg        *spvConfig
	HTTPClient HTTPClient
}

// NewClient create a new rest api client by providing functional config settings
//
// Example of usage :
//
//
//	client := spv.NewClient(
//		spv.WithBaseURL("localhost:5010"),
//		spv.WithVersion("v1"),
//		spv.WithUser("dev"),
//		spv.WithPassword("dev"),
//		spv.WithInsecure(),
//	)
//
// The full list of functional settings for a rest client are :
//
// To disable the TSL certificate check ( used in dev only )
//
//   WithInsecure()
//
// To set the base url of the server
//
//   WithBaseURL(url string)
//
// To set the version string of the rest api
//
//   WithVersion(v string)
//
// To set the user's name for basic authentification
//
//   WithUser(userName string)
//
// To set the user's password for the basic authentification
//
//   WithPassword(p string)
//
// To set the brearer token authentification (this will ignore the basic authentification if set)
//
// WithToken(t string)
func NewClient(opts ...SPVConfigFunc) *Client {

	// Start with the defaults then overwrite config with any set by user
	cfg := defaultSPVConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := http.Client{
		Timeout: time.Minute,
	}

	if cfg.insecure {
		httpClient.Transport = &http.Transport{
			// #nosec
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &Client{
		cfg:        cfg,
		HTTPClient: &httpClient,
	}
}

// errorResponse hold structure of error rest call
type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// successResponse hold structure of success rest call
type successResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// sendRequest send the http request and receive the response
func (c *Client) sendRequest(req *http.Request, out interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	if c.cfg.token == "" {
		req.SetBasicAuth(c.cfg.user, c.cfg.passwd)
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.token))
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}

		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	fullResponse := successResponse{
		Code: res.StatusCode,
		Data: out,
	}

	if out != nil {
		if err = json.NewDecoder(res.Body).Decode(&fullResponse.Data); err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Websocket client is listening to the stream of notifications, which notifies new messages for a specific channel.
// It does not receive the content of the message itself (even though, the notification itself is a text)
// User has to write an separate engine to pull the new (unread) messages content when it receive a notification.
// This can be easily done through the existing endpoint Messages provided in the rest api

// WSClient is the structure holding the
//    - websocket configuration
//    - websocket connection
//    - number of received notifications
type WSClient struct {
	mu      sync.Mutex
	cfg     *spvConfig
	ws      *ws.Conn
	close   chan bool
	started bool
}

// NewWSClient create a new connected websocket client by providing fuctional config settings.
// After being created (connected), the websocket client is ready to listen to new messages
//
// Example of usage :
//
//
//	ws := spv.NewWSClient(
//		spv.WithBaseURL("localhost:5010"),
//		spv.WithVersion("v1"),
//		spv.WithChannelID(channelid),
//		spv.WithToken(tok),
//		spv.WithInsecure(),
//		spv.WithWebsocketCallBack(PullUnreadMessages),
//	)
//
// The full list of functional settings for a websocket client are :
//
// To disable the TSL certificate check ( used in dev only )
//
//   WithInsecure()
//
// To set the base url of the server
//
//   WithBaseURL(url string)
//
// To set the version string of the server
//
//   WithVersion(v string)
//
// To set channel to be notified
//
//   WithChannelID(channelid string)
//
// To set the token that allow the socket connection
//
//   WithToken(tok string)
//
// To specify a callback function to process the notification
//
//   WithWebsocketCallBack(p PullUnreadMessages)
func NewWSClient(opts ...SPVConfigFunc) (*WSClient, error) {
	// Start with the defaults then overwrite config with any set by user
	cfg := defaultSPVConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	ws := &WSClient{
		cfg:   cfg,
		ws:    nil,
		close: make(chan bool),
	}

	err := ws.connectServer()

	if err != nil {
		return nil, err
	}

	return ws, nil
}

// urlPath return the path part of the connection URL
func (c *WSClient) urlPath() string {
	return path.Join(c.cfg.path, "/api", c.cfg.version, "/channel", c.cfg.channelID, "/notify")
}

// connectServer establish the connection to the server
// Return error if any
func (c *WSClient) connectServer() error {
	u := url.URL{
		Scheme: c.cfg.wsScheme(),
		Host:   c.cfg.baseURL,
		Path:   c.urlPath(),
	}

	q := u.Query()
	q.Set("token", c.cfg.token)
	u.RawQuery = q.Encode()

	d := ws.DefaultDialer
	if c.cfg.insecure {
		// #nosec
		d = &ws.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
			TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
		}
	}

	conn, httpRESP, err := d.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = httpRESP.Body.Close()
	}()

	c.ws = conn

	return nil
}

// Close stops reading any notification and closes the websocket
// Usually it is called from a separate goroutine
func (c *WSClient) Close() {
	if c.close == nil {
		return
	}
	c.close <- true
	close(c.close)
	c.close = nil
	_ = c.ws.Close()

}

// Run establishes the connection and start listening the notification stream
// process the notification if a callback is provided
func (c *WSClient) Run() {
	go func() {
		defer func() {
			_ = recover()
			c.Close()
		}()
		c.mu.Lock()
		c.started = true
		c.mu.Unlock()
		for {
			t, msg, err := c.ws.ReadMessage()
			if c.cfg.procces != nil {
				if err2 := c.cfg.procces(context.Background(), t, msg, err); err2 != nil {
					if errors.Is(err2, ErrWSClose{}) {
						return
					}
					c.cfg.errHandler(err2)
				}
			}
		}
	}()

	<-c.close
}
