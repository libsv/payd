package service

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type peerChannelsNotifySvc struct {
	cfg           *config.PeerChannels
	mu            sync.RWMutex
	pcSvc         payd.PeerChannelsService
	subscriptions map[string]*payd.PeerChannelSubscription
	handlers      map[payd.PeerChannelHandlerType]payd.PeerChannelsMessageHandler
}

func NewPeerChannelsNotifyService(cfg *config.PeerChannels, pcSvc payd.PeerChannelsService) payd.PeerChannelsNotifyService {
	return &peerChannelsNotifySvc{
		cfg:           cfg,
		pcSvc:         pcSvc,
		subscriptions: make(map[string]*payd.PeerChannelSubscription),
		handlers:      make(map[payd.PeerChannelHandlerType]payd.PeerChannelsMessageHandler),
	}
}

func (p *peerChannelsNotifySvc) RegisterHandler(ht payd.PeerChannelHandlerType, hdlr payd.PeerChannelsMessageHandler) payd.PeerChannelsNotifyService {
	p.handlers[ht] = hdlr
	return p
}

func (p *peerChannelsNotifySvc) Subscribe(ctx context.Context, channel *payd.PeerChannel) error {
	if _, ok := p.handlers[channel.Type]; !ok {
		return fmt.Errorf("unrecognised channel type '%s'", string(channel.Type))
	}

	u := url.URL{
		Scheme: "ws",
		Host:   p.cfg.Host,
		Path:   path.Join("/api/v1/channel", string(channel.ID), "/notify"),
	}
	q := u.Query()
	q.Set("token", channel.Token)
	u.RawQuery = q.Encode()

	ws, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if err == websocket.ErrBadHandshake {
			return errors.Wrapf(err, "notification subscription handshake failed %d", resp.StatusCode)
		}
	}
	defer resp.Body.Close()

	sub := payd.PeerChannelSubscription{
		ChannelID:   channel.ID,
		ChannelType: channel.Type,
		Token:       channel.Token,
		Conn:        ws,
	}

	func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.subscriptions[channel.ID] = &sub
	}()

	go p.listen(context.Background(), &sub)
	return nil
}

func (p *peerChannelsNotifySvc) listen(ctx context.Context, sub *payd.PeerChannelSubscription) {
	if err := func() error {
		defer sub.Conn.Close()
		_, msg, err := sub.Conn.ReadMessage()
		if err != nil {
			return err
		}

		log.Info().Msg(sub.ChannelID + " " + string(msg))
		return nil
	}(); err != nil {
		log.Error().Err(errors.WithStack(err))
	}

	log.Error().Err(errors.WithStack(p.handleNotification(ctx, sub)))
}

func (p *peerChannelsNotifySvc) handleNotification(ctx context.Context, sub *payd.PeerChannelSubscription) error {
	msgs, err := p.pcSvc.PeerChannelsMessage(ctx, &payd.PeerChannelMessageArgs{
		ChannelID: sub.ChannelID,
		Token:     sub.Token,
	})
	if err != nil {
		log.Error().Err(errors.WithStack(err))
		return err
	}

	hdlr := p.handlers[sub.ChannelType]
	finished, err := hdlr.HandlePeerChannelsMessage(ctx, msgs)
	if err != nil {
		log.Error().Err(errors.WithStack(err))
		return err
	}

	if finished {
		return p.pcSvc.CloseChannel(ctx, sub.ChannelID)
	}

	return nil
}

func (p *peerChannelsNotifySvc) cleanup(sub string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.subscriptions, sub)
}
