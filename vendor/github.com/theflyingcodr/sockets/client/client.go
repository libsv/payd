package client

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/internal"
)

type opts struct {
	reconnect         bool
	reconnectAttempts int
	reconnectTimeout  time.Duration
	writeTimeout      time.Duration
	pongWait          time.Duration
	maxMessageBytes   int64
}

func defaultOpts() *opts {
	o := &opts{
		reconnect:         false,
		reconnectAttempts: 3,
		reconnectTimeout:  30 * time.Second,
		writeTimeout:      2 * time.Second,
		pongWait:          60 * time.Second,
		maxMessageBytes:   512,
	}
	return o
}

// OptFunc defines a functional option to pass to the client at setup time.
type OptFunc func(c *opts)

// WithReconnect will enable reconnects from a client,
// in the event of a connection loss with a server the client
// will attempt to reconnect.
//
// Default values are to retry 3 times with a 30 second wait between retry.
func WithReconnect() OptFunc {
	return func(c *opts) {
		c.reconnect = true
	}
}

// WithReconnectAttempts will overwrite the default connection attempts of
// 3 with value attempts, when this value is exceeded the connection will
// cease to re-connect and exit.
func WithReconnectAttempts(attempts int) OptFunc {
	return func(c *opts) {
		c.reconnectAttempts = attempts
	}
}

// WithReconnectTimeout will overwrite the default timeout between reconnect
// attempts of 30 seconds with value t.
func WithReconnectTimeout(t time.Duration) OptFunc {
	return func(c *opts) {
		c.reconnectTimeout = t
	}
}

// WithInfiniteReconnect will make the client listen forever for the server to
// reconnect in the event of a connection loss.
func WithInfiniteReconnect() OptFunc {
	return func(c *opts) {
		c.reconnectAttempts = -1
	}
}

// WithWriteTimeout defines the timeout length that the client will wait before
// failing the write.
// Default is 60 seconds.
func WithWriteTimeout(t time.Duration) OptFunc {
	return func(c *opts) {
		c.writeTimeout = t
	}
}

// WithPongTimeout defines the wait time the client will wait for a pong response
// from the server.
// Default is 60 seconds.
func WithPongTimeout(t time.Duration) OptFunc {
	return func(c *opts) {
		c.pongWait = t
	}
}

// WithMaxMessageSize defines the maximum message size in bytes that
// the client will accept.
// Default is 512 bytes.
func WithMaxMessageSize(s int64) OptFunc {
	return func(c *opts) {
		c.maxMessageBytes = s
	}
}

// Client contains a socket client which connects to one or many servers
// and channels on those servers.
type Client struct {
	conn             map[string]*connection
	listeners        map[string]sockets.HandlerFunc
	middleware       []sockets.MiddlewareFunc
	errHandler       sockets.ClientErrorHandlerFunc
	serverErrHandler sockets.ClientErrorMsgHandlerFunc
	close            chan struct{}
	done             chan struct{}
	sender           chan sendMsg
	channelJoin      chan *connection
	channelLeave     chan string
	channelReconnect chan reconnectChannel
	channelChecker   chan internal.ChannelCheck
	join             chan joinSuccess
	opts             *opts
	sync.RWMutex
}

type joinSuccess struct {
	ChannelID string
	ClientID  string
}

// New will setup a new websocket client which will connect to the server
// at the provided uri.
func New(opts ...OptFunc) *Client {
	o := defaultOpts()
	for _, opt := range opts {
		opt(o)
	}

	cli := &Client{
		conn:             make(map[string]*connection),
		listeners:        make(map[string]sockets.HandlerFunc),
		middleware:       make([]sockets.MiddlewareFunc, 0),
		errHandler:       defaultErrorHandler,
		serverErrHandler: defaultErrorMsgHandler,

		close:            make(chan struct{}, 1),
		done:             make(chan struct{}, 1),
		sender:           make(chan sendMsg, 256),
		channelJoin:      make(chan *connection, 1),
		channelLeave:     make(chan string, 1),
		channelReconnect: make(chan reconnectChannel, 1),
		channelChecker:   make(chan internal.ChannelCheck, 256),
		join:             make(chan joinSuccess, 1),
		RWMutex:          sync.RWMutex{},
		opts:             o,
	}
	cli.RegisterListener(sockets.MessageJoinSuccess, cli.joinSuccess)
	cli.RegisterListener(sockets.MessageChannelExpired, cli.channelExpired)
	cli.RegisterListener(sockets.MessageChannelClosed, cli.channelClosed)
	go cli.channelManager()
	return cli
}

// WithJoinRoomSuccessListener will replace the default room join success handler
// with a custom one. This will allow the implementor to define their own logic
// after a room is joined.
func (c *Client) WithJoinRoomSuccessListener(l sockets.HandlerFunc) *Client {
	c.Lock()
	defer c.Unlock()
	c.listeners[sockets.MessageJoinSuccess] = l
	return c
}

// WithChannelExpiredListener will replace the default channel expired handler which simply
// prints a debug message.
//
// By adding your own you can handle the channel expiry in a custom manner.
func (c *Client) WithChannelExpiredListener(l sockets.HandlerFunc) *Client {
	c.Lock()
	defer c.Unlock()
	c.listeners[sockets.MessageChannelExpired] = l
	return c
}

// WithChannelClosedListener will replace the default channel closed handler which simply
// prints a debug message. This is raised when the server explicitly closes a channel.
//
// By adding your own you can handle the channel close in a custom manner.
func (c *Client) WithChannelClosedListener(l sockets.HandlerFunc) *Client {
	c.Lock()
	defer c.Unlock()
	c.listeners[sockets.MessageChannelExpired] = l
	return c
}

// WithMiddleware will append the middleware funcs to any already registered middleware functions.
// When adding middleware, it is recommended to always add a PanicHandler first as this will ensure your
// application has the best chance of recovering. There is a default panic handler available under sockets.PanicHandler.
func (c *Client) WithMiddleware(mws ...sockets.MiddlewareFunc) *Client {
	c.middleware = append(c.middleware, mws...)
	return c
}

// WithJoinRoomFailedListener will replace the default room join failed handler
// with a custom one. This will allow the implementor to define their own logic
// after a room is joined.
func (c *Client) WithJoinRoomFailedListener(l sockets.HandlerFunc) *Client {
	c.Lock()
	defer c.Unlock()
	c.listeners[sockets.MessageLeaveSuccess] = l
	return c
}

func (c *Client) listener(name string) sockets.HandlerFunc {
	c.RLock()
	defer c.RUnlock()
	return c.listeners[name]
}

// WithErrorHandler allows a user to overwrite the default error handler.
func (c *Client) WithErrorHandler(e sockets.ClientErrorHandlerFunc) *Client {
	c.Lock()
	defer c.Unlock()
	c.errHandler = e
	return c
}

// WithServerErrorHandler allows a user to overwrite the default server error handler.
//
// This handles ErrorMessage responses from a server in response to a client send.
func (c *Client) WithServerErrorHandler(e sockets.ClientErrorMsgHandlerFunc) *Client {
	c.Lock()
	defer c.Unlock()
	c.serverErrHandler = e
	return c
}

// Close will ensure the client is gracefully shut down.
func (c *Client) Close() {
	log.Debug().Msg("closing socket client")
	for _, conn := range c.conn {
		conn.close()
	}
	log.Info().Msg("socket client closed")
}

// JoinChannel will connect the client to the supplied host and channelID, returning an error if
// it cannot connect.
//
// If you need to authenticate with the server or send meta, add header/s.
func (c *Client) JoinChannel(host, channelID string, headers http.Header, params map[string]string) error {
	log.Info().Msgf("joining channel %s", channelID)
	u, err := url.Parse(fmt.Sprintf("%s/%s", host, channelID))
	if err != nil {
		return errors.Wrap(err, "failed to parse channel url")
	}
	if params != nil {
		q := u.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		u.RawQuery = q.Encode()
	}
	ws, resp, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	c.channelJoin <- &connection{
		url:       u.String(),
		channelID: channelID,
		ws:        ws,
		closer:    make(chan bool),
		done:      make(chan struct{}),
	}
	log.Info().Msgf("connected to channel %s", channelID)
	return nil
}

// LeaveChannel will disconnect a client from a channel.
func (c *Client) LeaveChannel(channelID string, headers http.Header) {
	c.channelLeave <- channelID
}

// HasChannel will check to see if a client is connected to a channel.
func (c *Client) HasChannel(channelID string) bool {
	log.Debug().Msgf("checking if channel %s exists", channelID)
	exists := make(chan bool)
	defer close(exists)

	c.channelChecker <- internal.ChannelCheck{
		ID:     channelID,
		Exists: exists,
	}

	result := <-exists
	log.Debug().Msgf("channel %s exists: %t", channelID, result)
	return result
}

func (c *Client) channelManager() {
	for {
		select {
		case msg, ok := <-c.sender:
			ch := c.conn[msg.m.ChannelID()]
			if ch == nil {
				continue
			}
			if !ok {
				_ = internal.Write(ch.ws, c.opts.writeTimeout, websocket.CloseMessage, []byte{})
				log.Debug().Msgf("closing connection for channelID %s", ch.channelID)
				return
			}
			msg.m.ClientID = ch.clientID
			if err := internal.WriteJSON(ch.ws, c.opts.writeTimeout, msg.m); err != nil {
				if msg.notify != nil {
					msg.notify <- err
				}
				continue
			}
			if msg.notify != nil {
				msg.notify <- nil
			}
		case join := <-c.channelJoin:
			join.ws.SetReadLimit(c.opts.maxMessageBytes)
			_ = join.ws.SetReadDeadline(time.Now().Add(c.opts.pongWait))
			join.ws.SetPongHandler(func(string) error { _ = join.ws.SetReadDeadline(time.Now().Add(c.opts.pongWait)); return nil })
			c.conn[join.channelID] = join
			go c.listen(join)
		case channelID := <-c.channelLeave:
			ch := c.conn[channelID]
			if ch == nil {
				continue
			}
			ch.close()
			<-ch.done
			delete(c.conn, channelID)
		case s := <-c.join:
			ch := c.conn[s.ChannelID]
			if ch == nil {
				continue
			}
			ch.clientID = s.ClientID
		case r := <-c.channelReconnect:
			ch := c.conn[r.channelID]
			if ch == nil {
				continue
			}
			ch.ws = r.conn
		case e := <-c.channelChecker:
			_, ok := c.conn[e.ID]
			e.Exists <- ok
		}
	}
}

type reconnectChannel struct {
	channelID string
	conn      *websocket.Conn
}

// Publish will broadcast a message to the server and wait for an error.
func (c *Client) Publish(req sockets.Request) error {
	if req.ChannelID == "" || req.MessageKey == "" {
		return errors.New("channelID and msgType required")
	}
	msg := sockets.NewMessage(req.MessageKey, "", req.ChannelID)
	if err := msg.WithBody(req.Body); err != nil {
		return err
	}
	msg.Headers = req.Headers
	err := make(chan error)
	defer close(err)
	c.sender <- sendMsg{
		m:      msg,
		notify: err,
	}
	return <-err
}

// RegisterListener will add a new listener to the client.
func (c *Client) RegisterListener(msgType string, fn sockets.HandlerFunc) {
	c.Lock()
	defer c.Unlock()
	c.listeners[msgType] = fn
}
