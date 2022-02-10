package service

import (
	"context"
	"time"

	"github.com/libsv/go-spvchannels"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/pkg/errors"
)

type peerChannelsSvc struct {
	cfg *config.PeerChannels
	str payd.PeerChannelsStore
}

// NewPeerChannelsSvc return a new peer channel service.
func NewPeerChannelsSvc(str payd.PeerChannelsStore, cfg *config.PeerChannels) payd.PeerChannelsService {
	return &peerChannelsSvc{
		cfg: cfg,
		str: str,
	}
}

func (p *peerChannelsSvc) PeerChannelCreate(ctx context.Context, req spvchannels.ChannelCreateRequest) (*payd.PeerChannel, error) {
	c := spvchannels.NewClient(
		spvchannels.WithUser("username"),
		spvchannels.WithPassword("password"),
		spvchannels.WithVersion("v1"),
		spvchannels.WithBaseURL(p.cfg.Host),
		spvchannels.WithNoTLS(),
	)
	ch, err := c.ChannelCreate(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error creating channel")
	}

	createdAt := time.Now()
	if err := p.str.PeerChannelCreate(ctx, &payd.PeerChannelCreateArgs{
		PeerChannelAccountID: req.AccountID,
		ChannelHost:          p.cfg.Host,
		ChannelID:            ch.ID,
		ChannelType:          payd.PeerChannelHandlerTypeProof,
		CreatedAt:            createdAt,
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to store peer channel information for channel %s", ch.ID)
	}

	return &payd.PeerChannel{
		ID:        ch.ID,
		Host:      p.cfg.Host,
		CreatedAt: createdAt,
		Type:      payd.PeerChannelHandlerTypeProof,
	}, nil
}

func (p peerChannelsSvc) PeerChannelAPITokensCreate(ctx context.Context, reqs ...*payd.PeerChannelAPITokenCreateArgs) ([]*spvchannels.TokenCreateReply, error) {
	c := spvchannels.NewClient(
		spvchannels.WithUser("username"),
		spvchannels.WithPassword("password"),
		spvchannels.WithVersion("v1"),
		spvchannels.WithBaseURL(p.cfg.Host),
		spvchannels.WithNoTLS(),
	)

	tokens := make([]*spvchannels.TokenCreateReply, 0)
	for _, req := range reqs {
		token, err := c.TokenCreate(ctx, req.Request)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create token")
		}

		if req.Persist {
			if err := p.str.PeerChannelAPITokenCreate(ctx, &payd.PeerChannelAPITokenStoreArgs{
				Role:                  req.Role,
				CanRead:               token.CanRead,
				CanWrite:              token.CanWrite,
				PeerChannelsChannelID: req.Request.ChannelID,
				Token:                 token.Token,
			}); err != nil {
				return nil, errors.Wrapf(err, "failed to store token %s", token.Token)
			}
		}

		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (p *peerChannelsSvc) PeerChannelsMessage(ctx context.Context, args *payd.PeerChannelMessageArgs) (spvchannels.MessagesReply, error) {
	c := spvchannels.NewClient(
		spvchannels.WithToken(args.Token),
		spvchannels.WithChannelID(args.ChannelID),
		spvchannels.WithVersion("v1"),
		spvchannels.WithBaseURL(p.cfg.Host),
		spvchannels.WithNoTLS(),
	)
	msgs, err := c.Messages(ctx, spvchannels.MessagesRequest{
		ChannelID: args.ChannelID,
		UnRead:    true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error reading messages from channel %s", args.ChannelID)
	}
	return msgs, nil
}

func (p *peerChannelsSvc) ActiveProofChannels(ctx context.Context) ([]payd.PeerChannel, error) {
	return p.str.PeerChannelsOpened(ctx, payd.PeerChannelHandlerTypeProof)
}

func (p *peerChannelsSvc) CloseChannel(ctx context.Context, channelID string) error {
	return p.str.PeerChannelCloseChannel(ctx, channelID)
}
