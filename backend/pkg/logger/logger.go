// Package logger constructs the application's zap logger.
//
// It lives under pkg/ rather than internal/ because it carries no domain
// knowledge and is the one component a sibling service could lift verbatim.
package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a logger appropriate to the environment.
//
// Development gets colour-coded, human-readable console output with caller
// links. Production gets structured JSON on stdout for the log pipeline to
// ingest — a human-formatted production log is a log nobody can query.
func New(level string, production bool) (*zap.Logger, error) {
	lvl, err := parseLevel(level)
	if err != nil {
		return nil, err
	}

	var cfg zap.Config
	if production {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	log, err := cfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, fmt.Errorf("logger: build: %w", err)
	}
	return log, nil
}

func parseLevel(s string) (zapcore.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "", "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("logger: unknown level %q", s)
	}
}

// Nop returns a logger that discards everything. Useful in tests that exercise
// code paths which log but whose output is not under test.
func Nop() *zap.Logger { return zap.NewNop() }
