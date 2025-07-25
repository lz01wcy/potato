package log

import (
	"github.com/asynkron/protoactor-go/actor"
	"github.com/lmittmann/tint"
	slogzap "github.com/samber/slog-zap/v2"
	"log/slog"
	"os"
	"time"
)

// enable JSON logging
func JsonLogging(system *actor.ActorSystem) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil)).
		With("system", system.ID)
}

// enable console logging
func ConsoleLogging(system *actor.ActorSystem) *slog.Logger {
	return slog.Default().
		With("system", system.ID)
}

// enable colored console logging
func ColoredConsoleLogging(system *actor.ActorSystem) *slog.Logger {
	return slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelWarn,
		TimeFormat: time.RFC3339,
		AddSource:  true,
	})).With("system", system.ID)
}

// enable Zap logging
func ZapAdapterLogging(system *actor.ActorSystem) *slog.Logger {
	zapLogger := Logger
	logger := slog.New(slogzap.Option{Level: slog.LevelInfo, Logger: zapLogger}.NewZapHandler())
	return logger.
		With("system", system.ID)
}
