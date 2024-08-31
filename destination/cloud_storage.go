package destination

import (
	"context"
	"io"

	"github.com/secmon-as-code/hatchery/pkg/metadata"
)

// CloudStorage is a destination that writes data to a Google cloud storage bucket.
type CloudStorage struct {
	bucket string
	prefix string
}

// NewCloudStorage creates a new CloudStorage destination.
func NewCloudStorage(bucket, prefix string) *CloudStorage {
	return &CloudStorage{
		bucket: bucket,
		prefix: prefix,
	}
}

func (c *CloudStorage) NewWriter(ctx context.Context, md metadata.MetaData) (io.WriteCloser, error) {
	// Open a new file in the cloud storage bucket.

	client, err := storage.NewClient(ctx)

	return nil, nil
}
