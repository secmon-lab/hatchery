package hatchery

import (
	"context"
	"io"

	"github.com/secmon-as-code/hatchery/pkg/metadata"
)

// Source is an interface that loads data from a source to a destination.
type Source interface {
	// Load loads data from a source to a destination. It should be called periodically to get data from the source. The interval of calling Load depends on command execution of hatchery.
	Load(ctx context.Context, p *Pipe) error
}

// Destination is an interface that writes data to data storage, messaging queue or something like that.
type Destination interface {
	NewWriter(ctx context.Context, md metadata.MetaData) (io.WriteCloser, error)
}
