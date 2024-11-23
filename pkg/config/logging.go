package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/m-mizutani/clog"
	"github.com/m-mizutani/goerr"
	"github.com/urfave/cli/v2"
)

type Logging struct {
	level  string
	format string
	out    string
}

func (x *Logging) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "log-format",
			Category:    "Logging",
			Aliases:     []string{"f"},
			EnvVars:     []string{"HATCHERY_LOG_FORMAT"},
			Usage:       "Log format (json, text)",
			Value:       "json",
			Destination: &x.format,
		},
		&cli.StringFlag{
			Name:        "log-level",
			Category:    "Logging",
			Aliases:     []string{"l"},
			EnvVars:     []string{"HATCHERY_LOG_LEVEL"},
			Usage:       "Log level (debug, info, warn, error)",
			Value:       "info",
			Destination: &x.level,
		},
		&cli.StringFlag{
			Name:        "log-out",
			Category:    "Logging",
			Aliases:     []string{"o"},
			EnvVars:     []string{"HATCHERY_LOG_OUT"},
			Usage:       "Log output (stdout or stderr)",
			Value:       "stdout",
			Destination: &x.out,
		},
	}
}

func (x *Logging) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("level", x.level),
		slog.String("format", x.format),
		slog.String("out", x.out),
	)
}

func (x *Logging) Build() (*slog.Logger, func(), error) {
	closer := func() {}
	var output io.Writer
	switch x.out {
	case "stdout", "-":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		f, err := os.OpenFile(filepath.Clean(x.out), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return nil, nil, goerr.Wrap(err, "Failed to open log file").With("path", x.out)
		}
		output = f
		closer = func() { f.Close() }
	}

	// Log level
	levelMap := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	level, ok := levelMap[x.level]
	if !ok {
		return nil, nil, goerr.New("Invalid log level").With("level", x.level)
	}

	// Log format
	var handler slog.Handler
	switch x.format {
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
		return nil, nil, goerr.New("Invalid log format").With("format", x.format)
	}

	return slog.New(handler), closer, nil

}
