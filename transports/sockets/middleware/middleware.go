package middleware

import (
	"context"

	"github.com/libsv/payd/config"
	tsoc "github.com/libsv/payd/transports/sockets"
	"github.com/theflyingcodr/sockets"
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
