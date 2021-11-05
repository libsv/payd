package server

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/theflyingcodr/sockets"
)

// RegisterDirectHandler will register handlers that respond ONLY to the client
// that sent them a message, no other clients will receive the notification.
func (s *SocketServer) RegisterDirectHandler(key string, fn sockets.HandlerFunc) *SocketServer {
	s.directListeners[key] = fn
	return s
}

// RegisterChannelHandler will add a handler that when sending a message will send to ALL clients
// connected to the channelID in the message.
func (s *SocketServer) RegisterChannelHandler(name string, fn sockets.HandlerFunc) *SocketServer {
	s.broadcastListeners[name] = fn
	return s
}

// defaultErrorHandler will simply log the error and then add some details
// to the message body before returning to the client the message was sent from.
func defaultErrorHandler(msg sockets.Message, e error) *sockets.ErrorMessage {
	log.Error().
		Str("id", msg.ID()).
		Str("trace", fmt.Sprintf("%v", e)).
		Str("msgType", msg.Key()).
		Err(e)

	return msg.ToError(sockets.ErrorDetail{
		Title:       "unexpected server error",
		Description: e.Error(),
		ErrCode:     "500",
	})
}
