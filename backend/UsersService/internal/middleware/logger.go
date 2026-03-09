package middleware

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger - инициализация logger
func InitLogger(logLevel string) (*zap.Logger, error) {
	var cfg zap.Config

	logLevel = strings.ToLower(strings.TrimSpace(logLevel))

	switch logLevel {
	case "release", "prod":
		cfg = zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

		// stacktrace только для panic / fatal
		cfg.DisableStacktrace = true

		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	case "development", "dev":
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)

		// без stacktrace на ErrorDetail/Warn
		cfg.DisableStacktrace = true

		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	default:
		return nil, fmt.Errorf("unknown logger logLevel: %s", logLevel)
	}

	logger, err := cfg.Build(zap.AddCaller())
	if err != nil {
		return nil, err
	}

	return logger, nil
}
