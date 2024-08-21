package destination

import (
	"context"
	"io"
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

func (c *CloudStorage) NewWriter(ctx context.Context) (io.WriteCloser, error) {
	// Write data to Google cloud storage bucket.
	return nil, nil
}
