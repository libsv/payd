package client

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/middleware"
)

type connection struct {
	url       string
	channelID string
	clientID  string
	closing   bool
	ws        *websocket.Conn
	closer    chan bool
	done      chan struct{}
}

func (c *connection) close() {
	log.Debug().Msgf("closing connection %s", c.clientID)
	if c.closing {
		return
	}
	c.closing = true
	c.closer <- true
	<-c.done
}

func (c *Client) listen(conn *connection) {
	channelID := conn.channelID
	url := conn.url

	go func() {
		defer conn.close()
		for {
			var body map[string]interface{}
			_, bb, err := conn.ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
					log.Error().Err(errors.WithStack(err)).Msg("received unknown read error")
					ws, ok := c.reconnect(conn.url)
					if !ok {
						log.Error().
							Msgf("failed to re-connect to %s after %d attempts, exiting conn", url, c.opts.reconnectAttempts)
						c.LeaveChannel(channelID, nil)
						return
					}
					conn.ws = ws
					log.Debug().Msg("reconnected to server")
					continue
				}
				log.Debug().
					Msgf("close message received for channelID '%s', closing connection", channelID)
				c.LeaveChannel(channelID, nil)
				return
			}

			if err := json.Unmarshal(bb, &body); err != nil {
				log.Err(err).Msg("error when reading message")
				continue
			}
			if body["type"] == sockets.MessageError {
				var errMsg sockets.ErrorMessage
				if err := json.Unmarshal(bb, &errMsg); err != nil {
					continue
				}
				c.serverErrHandler(errMsg)
				continue
			}
			var msg *sockets.Message
			if err := json.Unmarshal(bb, &msg); err != nil {
				log.Error().Err(err).Msg("unknown message type received")
				continue
			}
			ctx := context.Background()
			log.Debug().
				Str("channelID", msg.ChannelID()).
				Str("clientID", msg.ClientID).
				Str("type", msg.Key()).
				Msg("new message received")
			fn := c.listener(msg.Key())
			if fn == nil {
				log.Warn().Msgf("no handler found for message type '%s'", msg.Key())
				continue
			}
			// exec middleware and then handler.
			resp, err := middleware.ExecMiddlewareChain(fn, c.middleware)(ctx, msg)
			if err != nil {
				c.errHandler(errors.WithStack(err), msg)
				continue
			}
			if resp != nil {
				log.Debug().
					RawJSON("message body", resp.Body).
					Str("key", resp.Key()).
					Msg("sending message")
				c.sender <- sendMsg{
					m: resp,
				}
			}
		}
	}()
	<-conn.closer
	log.Debug().Msgf("closing channelID %s connection", channelID)
	if err := conn.ws.Close(); err != nil {
		log.Err(err).Msgf("error when closing channelID '%s' socket connection", channelID)
	}
	close(conn.done)
}

func (c *Client) reconnect(url string) (*websocket.Conn, bool) {
	i := 0
	for {
		i++
		time.Sleep(c.opts.reconnectTimeout)
		ws, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Err(err).Msgf("failed to reconnect to '%s' after '%d' attempts", url, i)
			if c.opts.reconnectAttempts != -1 && i > c.opts.reconnectAttempts {
				return nil, false
			}
			continue
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		return ws, true
	}
}

type sendMsg struct {
	m      *sockets.Message
	notify chan error
}
