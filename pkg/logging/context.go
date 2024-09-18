package logging

import (
	"context"
	"log/slog"
)

type ctxLoggerKey struct{}

// InjectLogger injects logger to context. It's used to inject mock logger for testing.
func InjectCtx(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

// FromCtx returns logger from context. If logger is not found, it returns default logger.
func FromCtx(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
