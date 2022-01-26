package payd

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/libsv/go-spvchannels"
)

// The types of message handlers.
const (
	PeerChannelHandlerTypeProof PeerChannelHandlerType = "proof"
)

// PeerChannelHandlerType the type of function which a peer channels message should map to.
type PeerChannelHandlerType string

// PeerChannelsNotifyService for interfacing with peer channel notifications.
type PeerChannelsNotifyService interface {
	RegisterHandler(ht PeerChannelHandlerType, hdlr PeerChannelsMessageHandler) PeerChannelsNotifyService
	Subscribe(ctx context.Context, args *PeerChannel) error
}

// PeerChannelsMessageHandler for handling peer channel messages.
type PeerChannelsMessageHandler interface {
	HandlePeerChannelsMessage(ctx context.Context, msgs spvchannels.MessagesReply) (bool, error)
}

// PeerChannelSubscription for subscribing to channel notifications.
type PeerChannelSubscription struct {
	ChannelID   string
	Token       string
	ChannelType PeerChannelHandlerType
	Conn        *websocket.Conn
}
