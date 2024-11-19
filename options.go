package hatchery

import (
	"log/slog"
)

// WithLogger is an option to set a logger to the hatchery. The logger is used to log messages from the hatchery. This option is prioritized over other settings (e.g. CLI option)
func WithLogger(logger *slog.Logger) Option {
	return func(h *Hatchery) {
		h.logger = logger
		h.loggerIsDefault = false
	}
}
