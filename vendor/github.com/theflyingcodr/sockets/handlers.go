package sockets

import (
	"context"
)

// Common messages.
const (
	MessageJoinSuccess    = "join.success"
	MessageLeaveSuccess   = "leave.success"
	MessageGetInfo        = "get.info"
	MessageInfo           = "info"
	MessageError          = "error"
	MessageChannelExpired = "channel.expired"
	MessageChannelClosed  = "channel.closed"
)

// HandlerFunc defines listeners on both the server and clients.
// When registered these will be triggered when a message is received matching the key.
type HandlerFunc func(ctx context.Context, msg *Message) (*Message, error)

// MiddlewareFunc defines a common middleware signature that can be used to create
// middleware that executes before the HandlerFunc.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// ServerErrorHandlerFunc defines an error handler for a server.
//
// The message can be handled, logged, and returned to clients.
type ServerErrorHandlerFunc func(msg *Message, e error) *ErrorMessage

// ClientErrorHandlerFunc is raised when the client itself returns an error
// when processing a message.
// It can be logged, retried etc.
type ClientErrorHandlerFunc func(err error, msg *Message)

// ClientErrorMsgHandlerFunc is triggered when a server returns an error
// message in response to a client message.
// It will contain the detail of the error and the client can decide to log,
// re-send or do something else.
type ClientErrorMsgHandlerFunc func(err ErrorMessage)

// ErrorDetail is returned as the message body in the event of an error.
type ErrorDetail struct {
	Title       string
	Description string
	ErrCode     string
}
