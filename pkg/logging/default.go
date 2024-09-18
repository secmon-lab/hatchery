package logging

import (
	"log/slog"
	"os"
	"sync"
)

var (
	defaultLogger *slog.Logger
	loggerMutex   sync.RWMutex
)

func init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})
	defaultLogger = slog.New(handler)
}

func Default() *slog.Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	return defaultLogger
}

func SetDefault(logger *slog.Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	defaultLogger = logger
}
