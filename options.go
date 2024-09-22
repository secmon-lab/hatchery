package hatchery

import (
	"log/slog"
)

// WithLogger is an option to set a logger to the hatchery. The logger is used to log messages from the hatchery. If the logger is nil, it uses the default logger.
func WithLogger(logger *slog.Logger) Option {
	return func(h *Hatchery) {
		h.logger = logger
	}
}
