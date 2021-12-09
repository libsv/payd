package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/middleware"
)

type opts struct {
	writeTimeout    time.Duration
	pongWait        time.Duration
	pingPeriod      time.Duration
	maxMessageBytes int64
	channelTimeout  time.Duration
}

func defaultOpts() *opts {
	o := &opts{
		writeTimeout:    2 * time.Second,
		pongWait:        60 * time.Second,
		maxMessageBytes: 512,
		channelTimeout:  time.Duration(-1), // no timeout
	}
	o.pingPeriod = (o.pongWait * 9) / 10
	return o
}

// OptFunc defines a functional option to pass to the server at setup time.
type OptFunc func(c *opts)

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

// WithPingPeriod will define the break between pings to the server.
// This should always be less than PongTimeout.
func WithPingPeriod(i time.Duration) OptFunc {
	return func(c *opts) {
		c.pingPeriod = i
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

// WithChannelTimeout will setup the server with a timeout that
// is set on each channel as it is created.
//
// Default is to NOT expire, adding this will cause the server to
// automatically close the channels and emit a channel.expired message to each client.
func WithChannelTimeout(t time.Duration) OptFunc {
	return func(c *opts) {
		c.channelTimeout = t
	}
}

// WithNoChannelTimeout will setup the server so it won't ever expire
// channels, leaving it up to clients to disconnect.
//
// This is the default option.
func WithNoChannelTimeout() OptFunc {
	return func(c *opts) {
		c.channelTimeout = time.Duration(-1)
	}
}

// SocketServer is a central point that connects peers together.
// It manages connections and channels as well as sending of messages
// to peers.
//
// It can have listeners setup both for channel broadcast and direct broadcast.
type SocketServer struct {
	// maps clientID to roomID for direct client messaging
	clientConnections  map[string]*connection
	channels           map[string]*channel
	waitingMsgs        *waitMessages
	broadcastListeners map[string]sockets.HandlerFunc
	directListeners    map[string]sockets.HandlerFunc

	middleware      []sockets.MiddlewareFunc
	errHandler      sockets.ServerErrorHandlerFunc
	unregister      chan unregister
	register        chan register
	channelSender   chan sender
	directSender    chan sender
	close           chan struct{}
	done            chan struct{}
	channelCloser   chan string
	opts            *opts
	onRegister      func(clientID, channelID string)
	onDeRegister    func(clientID, channelID string)
	onChannelClose  func(channelID string)
	onChannelCreate func(channelID string)
}

// New will setup and return a new instance of a SocketServer.
func New(opts ...OptFunc) *SocketServer {
	defaults := defaultOpts()

	for _, o := range opts {
		o(defaults)
	}

	s := &SocketServer{
		clientConnections:  make(map[string]*connection),
		channels:           make(map[string]*channel),
		waitingMsgs:        newWaitMessgaes(),
		broadcastListeners: make(map[string]sockets.HandlerFunc),
		directListeners:    make(map[string]sockets.HandlerFunc),
		middleware:         make([]sockets.MiddlewareFunc, 0),
		errHandler:         defaultErrorHandler,
		unregister:         make(chan unregister, 1),
		register:           make(chan register, 1),
		channelSender:      make(chan sender, 256),
		directSender:       make(chan sender, 256),
		close:              make(chan struct{}, 1),
		done:               make(chan struct{}, 1),
		opts:               defaults,
		onRegister:         func(clientID, channelID string) {},
		onDeRegister:       func(clientID, channelID string) {},
		onChannelClose:     func(channelID string) {},
		onChannelCreate:    func(channelID string) {},
	}
	go s.channelManager()
	return s
}

func (s *SocketServer) channelManager() {
	ticker := time.NewTicker(time.Minute)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-s.close:
			close(s.channelSender)
			close(s.directSender)
			log.Info().Msg("closing server")
			for _, c := range s.clientConnections {
				_ = c.ws.Close()
			}
			close(s.unregister)
			close(s.register)

			log.Info().Msg("connections terminated")
			s.done <- struct{}{}
			return
		case u := <-s.unregister:
			conn, ok := s.clientConnections[u.clientID]
			if ok && conn != nil {
				_ = conn.ws.Close()
			}
			delete(s.clientConnections, u.clientID)
			ch := s.channels[u.channelID]
			if ch == nil {
				continue
			}
			delete(ch.conns, u.clientID)
			if len(ch.conns) == 0 {
				delete(s.channels, u.channelID)
				s.onChannelClose(u.channelID)
			}
			s.onDeRegister(u.clientID, u.channelID)
		case u := <-s.register:
			s.clientConnections[u.clientID] = u.connection
			ch, ok := s.channels[u.channelID]
			if !ok {
				if s.opts.channelTimeout == -1 {
					ch = newChannel(u.channelID, time.Time{})
				} else {
					ch = newChannel(u.channelID, time.Now().Add(s.opts.channelTimeout).UTC())
				}
				s.channels[u.channelID] = ch
				s.onChannelCreate(u.channelID)
			}
			ch.conns[u.clientID] = u.connection
			u.registered <- nil
			s.onRegister(u.clientID, u.channelID)
		case m := <-s.channelSender:
			log.Debug().Msg("running channel sender")
			ch := s.channels[m.ID]
			log.Debug().Msg("looked up channel")
			if ch == nil {
				log.Debug().Msgf("channel %s is nil", m.ID)
				continue
			}
			log.Debug().Msg("channel not nil")
			for _, sub := range ch.conns {
				sub.send <- m.msg
			}
			log.Debug().Msg("sent to all connections")
			// clear buffer
			n := len(s.channelSender)
			log.Debug().Msgf("buffer to clear %d", n)
			for i := 0; i < n; i++ {
				send, ok := <-s.channelSender
				if !ok {
					log.Debug().Msgf("channel sender %s is empty", m.ID)
					continue
				}
				ch := s.channels[send.ID]
				if ch == nil {
					continue
				}
				wg := sync.WaitGroup{}
				for _, sub := range ch.conns {
					wg.Add(1)
					go func(s *connection, m interface{}) {
						s.send <- m
						wg.Done()
					}(sub, send.msg)
				}
				wg.Wait()
			}
			log.Debug().Msgf("cleared channel buffers %d", len(s.channelSender))
		case m := <-s.directSender:
			ch := s.clientConnections[m.ID]
			if ch == nil {
				continue
			}
			ch.send <- m.msg
			// clear buffer
			n := len(s.directSender)
			for i := 0; i < n; i++ {
				send := <-s.directSender
				ch := s.clientConnections[send.ID]
				if ch == nil {
					continue
				}
				go func() { ch.send <- m.msg }()
			}
		case <-ticker.C:
			for channelID, channel := range s.channels {
				if channel.expires.IsZero() { // doesn't expire
					continue
				}
				if channel.expires.UTC().After(time.Now().UTC()) { // not yet expired
					continue
				}
				log.Debug().
					Msgf("channel %s expired, closing connections", channelID)
				// expired
				clients := channel.expire()
				delete(s.channels, channelID)
				for _, client := range clients {
					s.onDeRegister(client, channelID)
					delete(s.clientConnections, client)
				}
				s.onChannelClose(channelID)
			}
		case channelID := <-s.channelCloser:
			ch := s.channels[channelID]
			if ch == nil {
				continue
			}
			clients := ch.close()
			for _, client := range clients {
				s.onDeRegister(client, channelID)
				delete(s.clientConnections, client)
			}
			delete(s.channels, channelID)
			s.onChannelClose(channelID)
		}
	}
}

// OnClientJoin is called when a client joins a channel.
func (s *SocketServer) OnClientJoin(fn func(clientID, channelID string)) {
	if fn != nil {
		s.onRegister = fn
	}
}

// OnClientLeave is called when a client leaves a channel.
func (s *SocketServer) OnClientLeave(fn func(clientID, channelID string)) {
	if fn != nil {
		s.onDeRegister = fn
	}
}

// OnChannelClose is called when all clients have left a channel and it is closed.
func (s *SocketServer) OnChannelClose(fn func(channelID string)) {
	if fn != nil {
		s.onChannelClose = fn
	}
}

// OnChannelCreate is called when a new channel is created.
func (s *SocketServer) OnChannelCreate(fn func(channelID string)) {
	if fn != nil {
		s.onChannelCreate = fn
	}
}

// Listen will start up a new listener for the received connection and channelID.
//
// This would be called after an Upgrade call in an http handler usually in a go routine.
func (s *SocketServer) Listen(conn *websocket.Conn, channelID string) error {
	if channelID == "" {
		return errors.New("channelID cannot be empty")
	}
	conn.SetReadLimit(s.opts.maxMessageBytes)
	_ = conn.SetReadDeadline(time.Now().Add(s.opts.pongWait))
	conn.SetPongHandler(func(string) error { _ = conn.SetReadDeadline(time.Now().Add(s.opts.pongWait)); return nil })

	clientID := uuid.NewString()
	log.Info().Msgf("receiving new connection with clientID %s", clientID)
	c := &connection{
		ws:       conn,
		send:     make(chan interface{}, 256),
		clientID: clientID,
		opts:     s.opts,
	}
	go c.writer()
	log.Debug().Msgf("adding clientID %s connection to channelID %s", clientID, channelID)
	registered := make(chan error)
	s.register <- register{
		channelID:  channelID,
		clientID:   clientID,
		connection: c,
		registered: registered,
	}
	if err := <-registered; err != nil {
		return fmt.Errorf("failed to join channel %w", err)
	}
	// send the client a join success message
	channel := s.channels[channelID]
	log.Debug().Msgf("client %s added to channelID %s", clientID, channelID)
	defer func() {
		s.unregisterClient(channelID, clientID)
		log.Debug().Msgf("removed clientID %s", clientID)
	}()
	joinMsg := sockets.NewMessage(sockets.MessageJoinSuccess, clientID, channelID)
	if !channel.expires.IsZero() { // add the expiry header
		joinMsg.Headers.Add(sockets.HeaderChannelExpiry, channel.expires.UTC().Format(time.RFC3339))
	}
	s.BroadcastDirect(clientID, sockets.NewMessage(sockets.MessageJoinSuccess, clientID, channelID))

	log.Info().Msgf("connection with clientID %s added, listening for messages", clientID)

	for {
		var m *sockets.Message
		if err := conn.ReadJSON(&m); err != nil {
			log.Debug().Msg("clsoe error received, handling")
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Error().Msgf("unexpected client %s close error: %v", clientID, err)
			} else {
				log.Info().Msgf("client %s closed connection, exiting listener", clientID)
			}
			s.unregisterClient(channelID, clientID)
			break
		}

		m.ClientID = clientID
		ctx := context.Background()
		log.Debug().Msg("message received")
		// if this is a wait message deliver it and move on
		if wm := s.waitingMsgs.message(m.CorrelationID); wm != nil {
			log.Debug().Msgf("wait message for correlationID %s executing", m.CorrelationID)
			wm.deliver(m)
			continue
		}

		hndlr, isDirect := s.handler(m.Key())
		if hndlr == nil {
			log.Debug().Msgf("no handler found for message %s", m.Key())
			continue
		}
		log.Debug().Msgf("executing handler for message %s", m.Key())
		resp, err := middleware.ExecMiddlewareChain(hndlr, s.middleware)(ctx, m)
		if err != nil {
			errMsg := s.errHandler(m, errors.WithStack(err))
			if errMsg == nil {
				continue
			}
			s.directSender <- sender{
				ID:  clientID,
				msg: errMsg,
			}
			continue
		}
		// no response, nothing to broadcast
		if resp == nil {
			log.Debug().Msgf("nothing to broadcast")
			continue
		}
		if isDirect {
			log.Debug().Msgf("sending direct message")
			s.directSender <- sender{
				ID:  resp.ClientID,
				msg: resp,
			}
			continue
		}
		log.Debug().Msgf("sending channel message")
		s.channelSender <- sender{
			ID:  resp.ChannelID(),
			msg: resp,
		}
		log.Debug().Msgf("channel message sent")
	}
	return nil
}

// CloseChannel will close a channel and close all connections within.
func (s *SocketServer) CloseChannel(channelID string) {
	s.channelCloser <- channelID
}

// WithErrorHandler can be used to overwrite the default error handler.
func (s *SocketServer) WithErrorHandler(fn sockets.ServerErrorHandlerFunc) *SocketServer {
	s.errHandler = fn
	return s
}

// handler will return a handler, checking for direct listeners first, if not found nil is returned.
func (s *SocketServer) handler(name string) (sockets.HandlerFunc, bool) {
	l, ok := s.directListeners[name]
	if ok {
		return l, true
	}
	return s.broadcastListeners[name], false
}

// Close should always be called in a defer to allow the server
// to gracefully shutdown and close underling connections.
func (s *SocketServer) Close() {
	s.close <- struct{}{}
	<-s.done
}

// Broadcast will send a message to a channel.
//
// This is used if a server event happens that needs to be sent to all clients
// without a message being sent first via a listener.
func (s *SocketServer) Broadcast(channelID string, msg *sockets.Message) {
	s.channelSender <- sender{
		ID:  channelID,
		msg: msg,
	}
}

// BroadcastDirect will send a message directly to a client.
//
// This is used if a server event happens that needs to be sent to a client
// without a message being sent first via a listener.
func (s *SocketServer) BroadcastDirect(clientID string, msg *sockets.Message) {
	s.directSender <- sender{
		ID:  clientID,
		msg: msg,
	}
}

// BroadcastAwait will send a broadcast to a channel and wait for a response,
// this will simply act on the first response to hit the server, if multiple peers respond, only the
// first will be returned.
//
// The function will return if a msg is returned OR an error is returned OR the ctx times out.
func (s *SocketServer) BroadcastAwait(ctx context.Context, channelID string, msg *sockets.Message) (*sockets.Message, error) {
	defer s.waitingMsgs.delete(msg.CorrelationID)
	wm := newWaitMessage()
	s.waitingMsgs.add(msg.CorrelationID, wm)
	s.Broadcast(channelID, msg)
	return wm.wait(ctx)
}

// WithMiddleware will append the middleware funcs to any already registered middleware functions.
// When adding middleware, it is recommended to always add a PanicHandler first as this will ensure your
// application has the best chance of recovering. There is a default panic handler available under sockets.PanicHandler.
func (s *SocketServer) WithMiddleware(mws ...sockets.MiddlewareFunc) *SocketServer {
	s.middleware = append(s.middleware, mws...)
	return s
}

func (s *SocketServer) unregisterClient(channelID, clientID string) {
	s.unregister <- unregister{channelID, clientID}
}

type register struct {
	channelID  string
	clientID   string
	connection *connection
	registered chan error
}

type unregister struct {
	channelID string
	clientID  string
}

type sender struct {
	ID  string
	msg interface{}
}
