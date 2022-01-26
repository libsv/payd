package payd

import (
	"context"
	"time"

	"github.com/libsv/go-spvchannels"
)

// PeerChannelsService a service for interacting with peer channels.
type PeerChannelsService interface {
	PeerChannelCreate(ctx context.Context, req spvchannels.ChannelCreateRequest) (*PeerChannel, error)
	PeerChannelAPITokensCreate(ctx context.Context, reqs ...*PeerChannelAPITokenCreateArgs) ([]*spvchannels.TokenCreateReply, error)
	PeerChannelsMessage(ctx context.Context, args *PeerChannelMessageArgs) (spvchannels.MessagesReply, error)
	ActiveProofChannels(ctx context.Context) ([]PeerChannel, error)
	CloseChannel(ctx context.Context, channelID string) error
}

// PeerChannelsStore for interfacing with a peer channel data store.
type PeerChannelsStore interface {
	PeerChannelAccount(ctx context.Context, args *PeerChannelIDArgs) (*PeerChannelAccount, error)
	PeerChannelCreate(ctx context.Context, args *PeerChannelCreateArgs) error
	PeerChannelCloseChannel(ctx context.Context, channelID string) error
	PeerChannelsOpened(ctx context.Context, channelType PeerChannelHandlerType) ([]PeerChannel, error)
	PeerChannelAPITokenCreate(ctx context.Context, args *PeerChannelAPITokenStoreArgs) error
	PeerChannelAPITokensCreate(ctx context.Context, args ...*PeerChannelAPITokenStoreArgs) error
}

// PeerChannel data.
type PeerChannel struct {
	ID        string                 `db:"channel_id"`
	Token     string                 `db:"tok"`
	Host      string                 `db:"channel_host"`
	CreatedAt time.Time              `db:"created_at"`
	Type      PeerChannelHandlerType `db:"channel_type"`
}

// PeerChannelIDArgs for getting a peerchannel account of a user.
type PeerChannelIDArgs struct {
	UserID int64
}

// PeerChannelAccount a peer channel account.
type PeerChannelAccount struct {
	ID       int64
	Username string
	Password string
}

// PeerChannelCreateArgs for creating a peer channel.
type PeerChannelCreateArgs struct {
	PeerChannelAccountID int64                  `db:"peerchannels_account_id"`
	ChannelType          PeerChannelHandlerType `db:"channel_type"`
	ChannelHost          string                 `db:"channel_host"`
	ChannelID            string                 `db:"channel_id"`
	CreatedAt            time.Time              `db:"created_at"`
}

// PeerChannelAPITokenCreateArgs for creating a peer channel token.
type PeerChannelAPITokenCreateArgs struct {
	Role    string
	Persist bool
	Request spvchannels.TokenCreateRequest
}

// PeerChannelAPITokenStoreArgs for storing a peer channel token.
type PeerChannelAPITokenStoreArgs struct {
	PeerChannelsChannelID string `db:"peerchannels_channel_id"`
	Token                 string `db:"tok"`
	Role                  string `db:"role"`
	CanRead               bool   `db:"can_read"`
	CanWrite              bool   `db:"can_write"`
}

// PeerChannelMessageArgs for quering a peer channel message.
type PeerChannelMessageArgs struct {
	ChannelID string
	Token     string
}
