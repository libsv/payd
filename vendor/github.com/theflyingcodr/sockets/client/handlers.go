package client

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/theflyingcodr/sockets"
)

func defaultErrorHandler(msg *sockets.ErrorMessage) {
	var err sockets.ErrorDetail
	_ = msg.Bind(&err)
	log.Error().Str("originKey", msg.Key).
		Str("correlationID", msg.CorrelationID).
		RawJSON("errorDetail", msg.ErrorBody).
		Str("channelID", msg.ChannelID).Msg("server error received")
}

func (c *Client) joinSuccess(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	log.Debug().Msgf("joined channel %s success", msg.ChannelID())
	c.join <- joinSuccess{
		ChannelID: msg.ChannelID(),
		ClientID:  msg.ClientID,
	}
	return msg.NoContent()
}
