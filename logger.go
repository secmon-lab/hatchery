package hatchery

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/m-mizutani/clog"
	"github.com/m-mizutani/goerr"
)

func buildLogger(logLevel, logFormat, logOut string) (*slog.Logger, error) {
	var output io.Writer
	switch logOut {
	case "stdout", "-":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		f, err := os.OpenFile(filepath.Clean(logOut), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return nil, goerr.Wrap(err, "Failed to open log file").With("path", logOut)
		}
		output = f
	}

	// Log level
	levelMap := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	level, ok := levelMap[logLevel]
	if !ok {
		return nil, goerr.New("Invalid log level").With("level", logLevel)
	}

	// Log format
	var handler slog.Handler
	switch logFormat {
	case "text":
		handler = clog.New(
			clog.WithWriter(output),
			clog.WithLevel(level),
			clog.WithSource(true),

			// clog.WithTimeFmt("2006-01-02 15:04:05"),
			clog.WithColorMap(&clog.ColorMap{
				Level: map[slog.Level]*color.Color{
					slog.LevelDebug: color.New(color.FgGreen, color.Bold),
					slog.LevelInfo:  color.New(color.FgCyan, color.Bold),
					slog.LevelWarn:  color.New(color.FgYellow, color.Bold),
					slog.LevelError: color.New(color.FgRed, color.Bold),
				},
				LevelDefault: color.New(color.FgBlue, color.Bold),
				Time:         color.New(color.FgWhite),
				Message:      color.New(color.FgHiWhite),
				AttrKey:      color.New(color.FgHiCyan),
				AttrValue:    color.New(color.FgHiWhite),
			}),
		)

	case "json":
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		})

	default:
		return nil, goerr.New("Invalid log format").With("format", logFormat)
	}

	return slog.New(handler), nil
}
