package middleware

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"

	"github.com/theflyingcodr/sockets"
)

// TimeoutConfig contains configuration settings for timeing out requests
// sent to listeners.
type TimeoutConfig struct {
	// Skipper, if true will skip setting timeouts for the request.
	Skipper func(msg *sockets.Message) bool
	Timeout time.Duration
}

func defaultTimeoutSkipper(msg *sockets.Message) bool {
	return false
}

// NewTimeoutConfig will setup a new timeout config with default values.
// This sets the skipper to always return false (don't skip any requests) and
// the default timeout of 30 seconds.
func NewTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		Skipper: defaultTimeoutSkipper,
		Timeout: 30 * time.Second,
	}
}

// Timeout can be used to set a timeout on requests sent to listeners.
func Timeout(cfg *TimeoutConfig) sockets.MiddlewareFunc {
	return func(next sockets.HandlerFunc) sockets.HandlerFunc {
		return func(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
			if cfg.Skipper(msg) {
				return next(ctx, msg)
			}
			c, cancel := context.WithTimeout(ctx, cfg.Timeout)
			defer cancel()
			return next(c, msg)
		}
	}
}

// PanicHandler will handle panics and log the error allowing the service to recover.
func PanicHandler(next sockets.HandlerFunc) sockets.HandlerFunc {
	return func(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(string(debug.Stack()))
				log.Error().Err(err.(error)).Stack().Msg("panic occurred")
			}
		}()
		return next(ctx, msg)
	}
}

// LoggerConfig holds configuration for a middleware based logger.
// This is used to output detail of the messages received.
type LoggerConfig struct {
	Skipper func(msg *sockets.Message) bool
}

// NewLoggerConfig will setup a new LoggerConfig with a default
// skipper.
func NewLoggerConfig() *LoggerConfig {
	return &LoggerConfig{Skipper: defaultLogSkipper}
}

// AddSkipper will add a custom skipper function to the
// logger, you may want to skip some messages but not all, perhaps by message key
// or only logging every n events.
func (l *LoggerConfig) AddSkipper(fn func(msg *sockets.Message) bool) {
	l.Skipper = fn
}

func defaultLogSkipper(msg *sockets.Message) bool {
	return false
}

// Logger will log the entire received message.
func Logger(cfg *LoggerConfig) sockets.MiddlewareFunc {
	return func(next sockets.HandlerFunc) sockets.HandlerFunc {
		return func(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
			if cfg.Skipper(msg) {
				return next(ctx, msg)
			}
			log.Info().
				Str("id", msg.ID()).
				Str("key", msg.Key()).
				Str("correlationID", msg.CorrelationID).
				Str("appID", msg.AppID).
				Str("userID", msg.UserID).
				Str("clientID", msg.ClientID).
				Str("channelID", msg.ChannelID()).
				Time("timestamp", msg.Timestamp()).
				Str("headers", fmt.Sprintf("%s", msg.Headers)).
				RawJSON("body", msg.Body).
				Msg("new message received")

			resp, err := next(ctx, msg)
			if err != nil {
				return resp, err
			}
			log.Info().
				Str("id", msg.ID()).
				Str("key", msg.Key()).
				Str("correlationID", msg.CorrelationID).
				Str("appID", msg.AppID).
				Str("userID", msg.UserID).
				Str("clientID", msg.ClientID).
				Str("channelID", msg.ChannelID()).
				Time("timestamp", msg.Timestamp()).
				Str("headers", fmt.Sprintf("%s", msg.Headers)).
				RawJSON("body", msg.Body).
				Msg("message response")
			return resp, err
		}
	}
}

// ExecMiddlewareChain builds the global middleware chain recursively, functions are first class.
func ExecMiddlewareChain(f sockets.HandlerFunc, m []sockets.MiddlewareFunc) sockets.HandlerFunc {
	// if our chain is done, use the original handlerfunc
	if len(m) == 0 {
		return f
	}
	// otherwise nest the handlerfuncs
	return m[0](ExecMiddlewareChain(f, m[1:]))
}

// Metrics contains basic example metrics that can be captured
// and pulled into prometheus.
//
// You would most likely implement your own version of this.
func Metrics() sockets.MiddlewareFunc {
	opsProcessed := promauto.NewCounter(prometheus.CounterOpts{
		Name: "received_messages",
		Help: "The total number of received messages",
	})

	return func(next sockets.HandlerFunc) sockets.HandlerFunc {
		return func(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
			opsProcessed.Inc()
			return next(ctx, msg)
		}
	}
}
