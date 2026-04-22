package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New создаёт zap.Logger с нужным уровнем
func New(level string) (*zap.Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	log, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}
	return log, nil
}
