package payd

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/libsv/go-spvchannels"
)

const (
	PeerChannelHandlerTypeProof PeerChannelHandlerType = "proof"
)

type PeerChannelHandlerType string

type PeerChannelsNotifyService interface {
	RegisterHandler(ht PeerChannelHandlerType, hdlr PeerChannelsMessageHandler) PeerChannelsNotifyService
	Subscribe(ctx context.Context, args *PeerChannel) error
}

type PeerChannelsMessageHandler interface {
	HandlePeerChannelsMessage(ctx context.Context, msgs spvchannels.MessagesReply) (bool, error)
}

type PeerChannelSubscription struct {
	ChannelID   string
	Token       string
	ChannelType PeerChannelHandlerType
	Conn        *websocket.Conn
}
