package client

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
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
	log.Debug().Msgf("closing connection %s", c.channelID)
	if c.closing {
		return
	}
	c.closing = true
	c.closer <- true
	<-c.done
}

func (c *Client) listen(client *connection) {
	channelID := client.channelID
	url := client.url

	go func() {
		defer client.close()
		for {
			var body map[string]interface{}
			msgType, bb, err := client.ws.ReadMessage()
			if msgType == websocket.CloseMessage {
				log.Info().
					Msgf("close message received for channelID '%s', closing connection", channelID)
				return
			}
			if err != nil {
				log.Err(err).Msg("error when reading message")

				if msgType == -1 {
					log.Info().Msg("lost connection to server, retrying")
					ws, ok := c.reconnect(client.url)
					if !ok {
						log.Error().
							Msgf("failed to re-connect to %s after %d attempts, exiting client", url, c.opts.reconnectAttempts)
						c.channelLeave <- client.channelID
						return
					}
					client.ws = ws
					log.Info().Msg("reconnected to server")
				}
				continue
			}
			if err := json.Unmarshal(bb, &body); err != nil {
				log.Err(err).Msg("error when reading message")
				continue
			}
			if body["type"] == sockets.MessageError {
				var errMsg *sockets.ErrorMessage
				if err := json.Unmarshal(bb, &errMsg); err != nil {
					continue
				}
				c.errHandler(errMsg)
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
				log.Info().Msgf("no handler found for message type '%s'", msg.Key())
				continue
			}
			// exec middleware and then handler.
			resp, err := middleware.ExecMiddlewareChain(fn, c.middleware)(ctx, msg)
			if err != nil {
				if resp != nil {
					c.errHandler(resp.ToError(err))
					continue
				}
				c.errHandler(msg.ToError(err))
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
	<-client.closer
	log.Debug().Msgf("closing channelID %s", channelID)
	if err := client.ws.Close(); err != nil {
		log.Err(err).Msgf("error when closing channelID '%s' socket connection", channelID)
	}
	close(client.done)
}

func (c *Client) reconnect(url string) (*websocket.Conn, bool) {
	i := 0
	for {
		i++
		time.Sleep(c.opts.reconnectTimeout)
		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Err(err).Msgf("failed to reconnect to '%s' after '%d' attempts", url, i)
			if c.opts.reconnectAttempts != -1 && i > c.opts.reconnectAttempts {
				return nil, false
			}
			continue
		}
		return ws, true
	}
}

type sendMsg struct {
	m      *sockets.Message
	notify chan error
}
