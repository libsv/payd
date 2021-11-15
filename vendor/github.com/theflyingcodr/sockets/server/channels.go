package server

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/internal"
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
	fmt.Println("total cons", len(c.conns))
	for clientID, conn := range c.conns {
		fmt.Println("sending expired message")
		_ = internal.WriteJSON(conn.ws, time.Minute, sockets.NewMessage(sockets.MessageChannelExpired, clientID, c.id))
		fmt.Println("sending close message")
		_ = conn.ws.WriteControl(websocket.CloseMessage, nil, time.Now().Add(60*time.Second))
		_ = conn.ws.Close()
		clients = append(clients, clientID)
		delete(c.conns, clientID)
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
