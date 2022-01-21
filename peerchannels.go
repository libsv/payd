package payd

import (
	"context"

	"github.com/libsv/go-spvchannels"
)

type PeerChannelsService interface {
	PeerChannelCreate(ctx context.Context, req spvchannels.ChannelCreateRequest) (*spvchannels.ChannelCreateReply, error)
	PeerChannelAPITokensCreate(ctx context.Context, reqs ...*PeerChannelAPITokenCreateArgs) ([]*spvchannels.TokenCreateReply, error)
	PeerChannelsMessage(ctx context.Context, args *PeerChannelMessageArgs) (spvchannels.MessagesReply, error)
	ActiveProofChannels(ctx context.Context) ([]PeerChannel, error)
	CloseChannel(ctx context.Context, channelID string) error
}

type PeerChannelsStore interface {
	PeerChannelAccount(ctx context.Context, args *PeerChannelIDArgs) (*PeerChannelAccount, error)
	PeerChannelCreate(ctx context.Context, args *PeerChannelCreateArgs) error
	PeerChannelCloseChannel(ctx context.Context, channelID string) error
	PeerChannelsOpened(ctx context.Context, channelType PeerChannelHandlerType) ([]PeerChannel, error)
	PeerChannelAPITokenCreate(ctx context.Context, args *PeerChannelAPITokenStoreArgs) error
	PeerChannelAPITokensCreate(ctx context.Context, args ...*PeerChannelAPITokenStoreArgs) error
}

type PeerChannelArgs struct {
	ID    string
	Token string
}

type PeerChannel struct {
	ID    string                 `db:"channel_id"`
	Token string                 `db:"token"`
	Host  string                 `db:"channel_host"`
	Type  PeerChannelHandlerType `db:"channel_type"`
}

type PeerChannelIDArgs struct {
	UserID int64
}

type PeerChannelAccount struct {
	ID       int64
	Username string
	Password string
}

type PeerChannelCreateArgs struct {
	PeerChannelAccountID int64                  `db:"peerchannels_account_id"`
	ChannelType          PeerChannelHandlerType `db:"channel_type"`
	ChannelHost          string                 `db:"channel_host"`
	ChannelID            string                 `db:"channel_id"`
}

type PeerChannelAPITokenCreateArgs struct {
	Role    string
	Persist bool
	Request spvchannels.TokenCreateRequest
}

type PeerChannelAPITokenStoreArgs struct {
	PeerChannelsChannelID string `db:"peerchannels_channel_id"`
	Token                 string `db:"token"`
	Role                  string `db:"role"`
	CanRead               bool   `db:"can_read"`
	CanWrite              bool   `db:"can_write"`
}

type PeerChannelMessageArgs struct {
	ChannelID string
	Token     string
}
