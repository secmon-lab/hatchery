package safe

import (
	"context"
	"io"

	"github.com/m-mizutani/goerr"
	"github.com/secmon-lab/hatchery/pkg/logging"
)

func CloseReader(ctx context.Context, r io.ReadCloser) {
	if r == nil {
		return
	}
	if err := r.Close(); err != nil {
		logger := logging.FromCtx(ctx)
		logger.Error("failed to close reader", "err", goerr.Wrap(err))
	}
}
