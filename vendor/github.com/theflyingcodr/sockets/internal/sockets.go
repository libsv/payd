package internal

import (
	"time"

	"github.com/gorilla/websocket"
)

// Write writes a message with the given message type and payload.
func Write(ws *websocket.Conn, timeout time.Duration, mt int, payload []byte) error {
	_ = ws.SetWriteDeadline(time.Now().Add(timeout))
	return ws.WriteMessage(mt, payload)
}

// WriteJSON writes a message with the given message type and payload.
func WriteJSON(ws *websocket.Conn, timeout time.Duration, payload interface{}) error {
	_ = ws.SetWriteDeadline(time.Now().Add(timeout))
	return ws.WriteJSON(payload)
}

// ChannelCheck for sending channel check messages.
type ChannelCheck struct {
	ID     string
	Exists chan bool
}
