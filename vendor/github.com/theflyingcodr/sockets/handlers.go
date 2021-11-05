package sockets

import (
	"context"
)

// Common messages.
const (
	MessageJoinSuccess  = "join.success"
	MessageLeaveSuccess = "leave.success"
	MessageGetInfo      = "get.info"
	MessageInfo         = "info"
	MessageError        = "error"
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
type ServerErrorHandlerFunc func(msg Message, e error) *ErrorMessage

// ClientErrorHandlerFunc defines the client side error handler. This is triggered
// when a listener returns an error and allows a global way of managing errors.
// This is executed after middleware.
type ClientErrorHandlerFunc func(msg *ErrorMessage)

// ErrorDetail is returned as the message body in the event of an error.
type ErrorDetail struct {
	Title       string
	Description string
	ErrCode     string
}
