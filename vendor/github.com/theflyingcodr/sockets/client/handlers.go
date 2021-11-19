package client

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/theflyingcodr/sockets"
)

// defaultErrorHandler will receive an error and the message that triggered the error.
//
// You can inspect the error, log it etc or retry handling the message.
func defaultErrorHandler(err error, msg *sockets.Message) {
	if err == nil {
		return
	}
	if msg == nil {
		log.Error().Err(err).
			Str("reason", err.Error()).
			Str("trace", fmt.Sprintf("%v", err)).
			Msg("unexpected client error received")
		return
	}
	log.Error().Err(err).
		Str("reason", err.Error()).
		Str("trace", fmt.Sprintf("%v", err)).
		Str("originKey", msg.Key()).
		Str("correlationID", msg.CorrelationID).
		Str("channelID", msg.ChannelID()).
		Msg("unexpected client error message received")
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

func (c *Client) channelExpired(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	log.Debug().Msgf("channel expired %s", msg.ChannelID())
	c.LeaveChannel(msg.ChannelID(), nil)
	return msg.NoContent()
}

func (c *Client) channelClosed(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	log.Debug().Msgf("channel closed %s", msg.ChannelID())
	c.LeaveChannel(msg.ChannelID(), nil)
	return msg.NoContent()
}
