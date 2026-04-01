package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new structured logger. In production it emits JSON; in
// development it uses the human-readable console encoder.
func New(env string) (*zap.Logger, error) {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return cfg.Build()
}

// Must is like New but panics on error — suitable for use at program start.
func Must(env string) *zap.Logger {
	l, err := New(env)
	if err != nil {
		panic(err)
	}
	return l
}
