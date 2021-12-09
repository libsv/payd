package server

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/theflyingcodr/sockets/internal"
)

type connection struct {
	ws       *websocket.Conn
	send     chan interface{}
	clientID string
	opts     *opts
}

// writer sends messages from the server to the websocket connection.
func (c *connection) writer() {
	ticker := time.NewTicker(c.opts.pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.ws.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				_ = internal.Write(c.ws, c.opts.writeTimeout, websocket.CloseMessage, []byte{})
				log.Debug().Msgf("closing connection for clientID %s", c.clientID)
				return
			}
			if err := internal.WriteJSON(c.ws, c.opts.writeTimeout, msg); err != nil {
				log.Err(err)
				return
			}
			n := len(c.send)
			for i := 0; i < n; i++ {
				msg, ok = <-c.send
				if !ok {
					_ = internal.Write(c.ws, c.opts.writeTimeout, websocket.CloseMessage, []byte{})
					log.Debug().Msgf("closing connection for clientID %s", c.clientID)
					return
				}
				go func(m interface{}) {
					if err := internal.WriteJSON(c.ws, c.opts.writeTimeout, m); err != nil {
						log.Err(err)
						return
					}
				}(msg)
			}
		case <-ticker.C:
			if err := internal.Write(c.ws, c.opts.writeTimeout, websocket.PingMessage, []byte{}); err != nil {
				log.Err(err)
				return
			}
		}
	}
}
