package client

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/theflyingcodr/sockets"
)

// defaultErrorHandler will receive an error and the message that triggered the error.
//
// You can inspect the error, log it etc or retry handling the message.
func defaultErrorHandler(err error, msg *sockets.Message) {
	log.Error().Err(err).
		Str("originKey", msg.Key()).
		Str("correlationID", msg.CorrelationID).
		Str("channelID", msg.ChannelID()).
		Msg("client error received")
}

// defaultErrorMsgHandler will receive an error and the message that triggered the error.
//
// You can inspect the error, log it etc or retry handling the message.
func defaultErrorMsgHandler(errMsg sockets.ErrorMessage) {
	var err sockets.ErrorDetail
	_ = errMsg.Bind(&err)
	log.Error().
		RawJSON("errorDetail", errMsg.ErrorBody).
		Str("originKey", errMsg.OriginKey).
		Str("correlationID", errMsg.CorrelationID).
		Str("channelID", errMsg.ChannelID).
		RawJSON("originBody", errMsg.OriginBody).
		Msg("server error received")
}

func (c *Client) joinSuccess(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	log.Debug().Msgf("joined channel %s success", msg.ChannelID())
	c.join <- joinSuccess{
		ChannelID: msg.ChannelID(),
		ClientID:  msg.ClientID,
	}
	return msg.NoContent()
}

func channelExpired(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	log.Debug().Msgf("channel expired %s", msg.ChannelID())
	return msg.NoContent()
}

func channelClosed(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	log.Debug().Msgf("channel closed %s", msg.ChannelID())
	return msg.NoContent()
}
