package middleware

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/theflyingcodr/sockets"

	"github.com/libsv/payd/config"
	tsoc "github.com/libsv/payd/transports/sockets"
)

// IgnoreMyMessages will drop any messages received on the socket that are sent
// by this client instance.
func IgnoreMyMessages(cfg *config.Socket) sockets.MiddlewareFunc {
	return func(next sockets.HandlerFunc) sockets.HandlerFunc {
		return func(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
			// we set an origin id when sending, this is the server random identifier a
			// client can use to identify messages sent my itself.
			// ignore the messages from our identifier.
			if msg.Headers.Get(tsoc.HeaderOrigin) == cfg.ClientIdentifier {
				return msg.NoContent()
			}
			resp, err := next(ctx, msg)
			if err != nil {
				return nil, err
			}
			if resp != nil {
				resp.Headers.Add(tsoc.HeaderOrigin, cfg.ClientIdentifier)
			}
			return resp, err
		}
	}
}

// WithAppIDPayD will add 'payd' as the appID for all outgoing messages.
func WithAppIDPayD() sockets.MiddlewareFunc {
	return func(next sockets.HandlerFunc) sockets.HandlerFunc {
		return func(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
			resp, err := next(ctx, msg)
			if err != nil {
				return nil, err
			}
			if resp != nil {
				resp.AppID = "payd"
			}
			return resp, err
		}
	}
}

// ErrorHandler will receive an error and the message that triggered the error.
//
// You can inspect the error, log it etc and if you want to send the error to
// the sender for them to also handle, return a struct with error details.
// If the return is nil no message will be sent back to the sender.
func ErrorHandler(err error, msg *sockets.Message) {
	log.Error().Err(err).
		Str("originKey", msg.Key()).
		Str("correlationID", msg.CorrelationID).
		Str("channelID", msg.ChannelID()).
		Msg("server error received")
}

// ErrorHandler will receive an error and the message that triggered the error.
//
// You can inspect the error, log it etc and if you want to send the error to
// the sender for them to also handle, return a struct with error details.
// If the return is nil no message will be sent back to the sender.
func ErrorMsgHandler(errMsg sockets.ErrorMessage) {
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
