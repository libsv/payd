package log

import (
	"github.com/libsv/payd/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger provides a basic logging definition that can be implemented
// by various log implementations and then passed as a dependency to
// code that requires logging.
type Logger interface {
	Info(s string)
	Infof(s string, a ...interface{})
	Debug(s string)
	Debugf(s string, a ...interface{})
	Warn(s string)
	Warnf(s string, a ...interface{})
	Error(err error, s string)
	Errorf(err error, s string, a ...interface{})
	Fatal(err error, s string)
	Fatalf(err error, s string, a ...interface{})
}

// Noop does nothing.
type Noop struct {
}

// Info writes an info level log.
func (n Noop) Info(s string) {}

// Infof writes an info level log with args.
func (n Noop) Infof(s string, a ...interface{}) {}

// Warn writes a warning level log.
func (n Noop) Warn(s string) {}

// Warnf writes a warning level log.
func (n Noop) Warnf(s string, a ...interface{}) {}

// Debug writes a debug level log.
func (n Noop) Debug(s string) {}

// Debugf writes a debug level log.
func (n Noop) Debugf(s string, a ...interface{}) {}

// Error writes an error log.
func (n Noop) Error(err error, s string) {}

// Errorf writes an error log.
func (n Noop) Errorf(err error, s string, a ...interface{}) {}

// Fatal writes a fatal log which will immediately terminate the program.
func (n Noop) Fatal(err error, s string) {}

// Fatalf writes a fatal log which will immediately terminate the program.
func (n Noop) Fatalf(err error, s string, a ...interface{}) {}

// Zero implements the Logger interface using zerolog.
type Zero struct {
}

// NewZero will create and return a new log using zero.
func NewZero(cfg *config.Logging) *Zero {
	lvl, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		log.Fatal().Msgf("failed to parse log level %s", err.Error())
	}
	zerolog.SetGlobalLevel(lvl)
	return &Zero{}
}

// Info writes an info level log.
func (z Zero) Info(s string) {
	log.Info().Msg(s)
}

// Infof writes an info level log with args.
func (z Zero) Infof(s string, a ...interface{}) {
	log.Info().Msgf(s, a...)
}

// Warn writes a warning level log.
func (z Zero) Warn(s string) {
	log.Warn().Msg(s)
}

// Warnf writes a warning level log.
func (z Zero) Warnf(s string, a ...interface{}) {
	log.Warn().Msgf(s, a...)
}

// Debug writes a debug level log.
func (z Zero) Debug(s string) {
	log.Debug().Msg(s)
}

// Debugf writes a debug level log.
func (z Zero) Debugf(s string, a ...interface{}) {
	log.Debug().Msgf(s, a...)
}

// Error writes an error log.
func (z Zero) Error(err error, s string) {
	log.Error().Err(errors.WithStack(err)).Msg(s)
}

// Errorf writes an error log.
func (z Zero) Errorf(err error, s string, a ...interface{}) {
	log.Error().Err(errors.WithStack(err)).Msgf(s, a...)
}

// Fatal writes a fatal log which will immediately terminate the program.
func (z Zero) Fatal(err error, s string) {
	log.Fatal().Err(errors.WithStack(err)).Msg(s)
}

// Fatalf writes a fatal log which will immediately terminate the program.
func (z Zero) Fatalf(err error, s string, a ...interface{}) {
	log.Fatal().Err(errors.WithStack(err)).Msgf(s, a...)
}
