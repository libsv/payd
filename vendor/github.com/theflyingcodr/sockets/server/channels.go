package server

import (
	"time"

	"github.com/theflyingcodr/sockets"
)

type channel struct {
	id      string
	conns   map[string]*connection
	expires time.Time
}

func newChannel(id string, expires time.Time) *channel {
	r := &channel{
		id:      id,
		conns:   make(map[string]*connection),
		expires: expires,
	}
	return r
}

// expire will expire all connections in the channel and return a list of affected clientIDs.
func (c *channel) expire() []string {
	clients := make([]string, 0)
	for clientID, conn := range c.conns {
		_ = conn.ws.WriteJSON(sockets.NewMessage(sockets.MessageChannelExpired, clientID, c.id))
		_ = conn.ws.Close()
		clients = append(clients, clientID)
	}
	return clients
}

// close will close all connections in the channel and return a list of affected clientIDs.
func (c *channel) close() []string {
	clients := make([]string, 0)
	for clientID, conn := range c.conns {
		_ = conn.ws.WriteJSON(sockets.NewMessage(sockets.MessageChannelExpired, clientID, c.id))
		_ = conn.ws.Close()
		clients = append(clients, clientID)
	}
	return clients
}
