package sockets

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/google/uuid"
)

// Client can be used to implement a client which will send and listen
// to messages on channels.
type Client interface {
	WithJoinRoomSuccessListener(l HandlerFunc) Client
	WithMiddleware(mws ...MiddlewareFunc) Client
	WithJoinRoomFailedListener(l HandlerFunc) Client
	WithErrorHandler(e ClientErrorHandlerFunc) Client
	Close()
	JoinChannel(host, channelID string, headers http.Header) error
	LeaveChannel(channelID string, headers http.Header)
	RegisterListener(msgType string, fn HandlerFunc) Client
}

// Request is used to send a message to a channel with a specific key.
type Request struct {
	ChannelID  string
	MessageKey string
	Body       interface{}
	Headers    http.Header
}

// Publisher is used by clients to send messages to a server.
type Publisher interface {
	Publish(req Request) error
}

// Message is the underlying message type used by the protocol to
// transmit metadata and the message bodies.
type Message struct {
	CorrelationID string
	AppID         string
	UserID        string
	Expiration    *time.Time
	Body          json.RawMessage
	id            string
	channelID     string
	timestamp     time.Time
	key           string
	Headers       http.Header
	ClientID      string
}

type messageJSON struct {
	CorrelationID string          `json:"correlationId"`
	AppID         string          `json:"appId"`
	ClientID      string          `json:"clientID"`
	UserID        string          `json:"userId"`
	Expiration    *time.Time      `json:"expiration"`
	Body          json.RawMessage `json:"body"`
	ID            string          `json:"messageId"`
	ChannelID     string          `json:"channelId"`
	Timestamp     time.Time       `json:"timestamp"`
	Key           string          `json:"type"`
	Headers       http.Header     `json:"headers"`
}

// MarshalJSON implements the json Marshaller.
func (m *Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(messageJSON{
		CorrelationID: m.CorrelationID,
		AppID:         m.AppID,
		ClientID:      m.ClientID,
		UserID:        m.UserID,
		Expiration:    m.Expiration,
		Body:          m.Body,
		ID:            m.id,
		ChannelID:     m.channelID,
		Timestamp:     m.timestamp,
		Key:           m.key,
		Headers:       m.Headers,
	})
}

// UnmarshalJSON implements the json unmarshaler.
func (m *Message) UnmarshalJSON(bb []byte) error {
	var j *messageJSON
	if err := json.Unmarshal(bb, &j); err != nil {
		return err
	}

	m.CorrelationID = j.CorrelationID
	m.AppID = j.AppID
	m.UserID = j.UserID
	m.Expiration = j.Expiration
	m.Body = j.Body
	m.id = j.ID
	m.channelID = j.ChannelID
	m.timestamp = j.Timestamp
	m.key = j.Key
	m.Headers = j.Headers
	m.ClientID = j.ClientID

	return nil
}

// ID returns the message unique identifier.
func (m *Message) ID() string {
	return m.id
}

// Timestamp returns the message created at date.
func (m *Message) Timestamp() time.Time {
	return m.timestamp
}

// ChannelID returns the message channelID.
func (m *Message) ChannelID() string {
	return m.channelID
}

// Key returns the message key.
func (m *Message) Key() string {
	return m.key
}

// NewFrom will take a copy of msg and return a new message from it.
//
// You can then add a body using the WithBody func and add headers etc.
func (m *Message) NewFrom(key string) *Message {
	msg := NewMessage(key, m.ClientID, m.channelID)
	msg.Expiration = m.Expiration
	msg.UserID = m.UserID
	msg.AppID = m.AppID
	msg.CorrelationID = m.CorrelationID
	return msg
}

// NewMessage will create a new message setting.
func NewMessage(msgType, clientID, channelID string) *Message {
	return &Message{
		id:        uuid.NewString(),
		timestamp: time.Now().UTC(),
		key:       msgType,
		Headers:   http.Header{},
		channelID: channelID,
		ClientID:  clientID,
	}
}

// Bind will map the body to v.
func (m Message) Bind(v interface{}) error {
	if m.Body == nil {
		return nil
	}
	// TODO - handle different binding ie params, headers xml etc
	return json.Unmarshal(m.Body, &v)
}

// WithBody will serialise the value v into the message body.
func (m *Message) WithBody(v interface{}) error {
	if isNil(v) {
		return nil
	}
	bb, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.Body = bb
	return nil
}

// NoContent is a helper that can be used to return an empty message from a listener.
func (m *Message) NoContent() (*Message, error) {
	return nil, nil
}

// DirectBroadcaster is used to send a message directly to a client.
type DirectBroadcaster interface {
	BroadcastDirect(clientID string, msg *Message)
}

// ChannelBroadcaster is used to send a message to all clients connected to a channel.
type ChannelBroadcaster interface {
	Broadcast(channelID string, msg *Message)
}

// ErrorMessage is a message type returned on error, it contains
// a copy of the original key and body but will have a key of "error"
// allowing this to be checked for in error handlers.
// ErrorBody is optional but if supplied can be marhsaled to a struct using Bind()
// to get the detail of the error.
type ErrorMessage struct {
	CorrelationID string          `json:"correlationId"`
	AppID         string          `json:"appId"`
	UserID        string          `json:"userId"`
	Key           string          `json:"type"`
	OriginKey     string          `json:"originType"`
	OriginBody    json.RawMessage `json:"originBody"`
	ErrorBody     json.RawMessage `json:"errorBody"`
	ChannelID     string          `json:"channelId"`
	ClientID      string          `json:"clientId"`
	Headers       http.Header     `json:"headers"`
}

// ToError will convert a message to an ErrorMessage with an optional
// err struct supplied containing additional details for the error.
//
// There is a default sockets.ErrorDetail struct available, or you can define your own.
func (m *Message) ToError(err interface{}) *ErrorMessage {
	var bb []byte
	if !isNil(err) {
		bb, _ = json.Marshal(err)
	}
	e := &ErrorMessage{
		CorrelationID: m.CorrelationID,
		AppID:         m.AppID,
		UserID:        m.UserID,
		Key:           MessageError,
		OriginKey:     m.key,
		OriginBody:    m.Body,
		ErrorBody:     bb,
		ChannelID:     m.channelID,
		ClientID:      m.ClientID,
		Headers:       m.Headers,
	}
	return e
}

// Bind will decode the error message body to v.
// This message will usually contain further error data and can be specified by the server.
func (e *ErrorMessage) Bind(v interface{}) error {
	if e.ErrorBody == nil {
		return nil
	}
	return json.Unmarshal(e.ErrorBody, &v)
}

// BindOriginBody can be used to decode the body for the original message that triggered
// this error, useful for replaying.
func (e *ErrorMessage) BindOriginBody(v interface{}) error {
	if e.OriginBody == nil {
		return nil
	}
	return json.Unmarshal(e.OriginBody, &v)
}

// isNil safely checks an interface for nil.
func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	// nolint: exhaustive // don't catering for everything... yet
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
