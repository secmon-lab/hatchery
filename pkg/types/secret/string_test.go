package secret_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/secmon-as-code/hatchery/pkg/types/secret"
)

func TestString(t *testing.T) {
	s := secret.NewString("blue")

	t.Run("text logger", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{}))

		logger.Info("test", "secret", s)
		gt.S(t, buf.String()).NotContains("blue")
	})

	t.Run("json logger", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{}))

		logger.Info("test", "secret", s)
		gt.S(t, buf.String()).NotContains("blue")
	})

	t.Run("Println", func(t *testing.T) {
		buf := &bytes.Buffer{}
		fmt.Fprintln(buf, "test", "secret", s)
		gt.S(t, buf.String()).NotContains("blue")
	})

	t.Run("Printf %s", func(t *testing.T) {
		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "test %s %s", "secret", s)
		gt.S(t, buf.String()).NotContains("blue")
	})

	t.Run("Printf %v", func(t *testing.T) {
		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "test %v %v", "secret", s)
		gt.S(t, buf.String()).NotContains("blue")
	})
}
