package destination

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/m-mizutani/goerr"
	"github.com/secmon-as-code/hatchery/pkg/metadata"
)

// Client is a destination that writes data to a Google cloud storage bucket.
type Client struct {
	bucket string
	prefix string
	gzip   bool
}

func (c *Client) Bucket() string { return c.bucket }
func (c *Client) Prefix() string { return c.prefix }
func (c *Client) Gzip() bool     { return c.gzip }

// New creates a new Client destination.
func New(bucket string, options ...Option) *Client {
	c := &Client{
		bucket: bucket,
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

type Option func(*Client) error

// WithPrefix sets a prefix for object names in the bucket.
func WithPrefix(prefix string) Option {
	return func(c *Client) error {
		c.prefix = prefix
		return nil
	}
}

// WithGzip sets a flag to compress data with gzip.
func WithGzip(gzip bool) Option {
	return func(c *Client) error {
		c.gzip = gzip
		return nil
	}
}

// NewWriter creates a new writer to write data to the cloud storage bucket.
func (c *Client) NewWriter(ctx context.Context, md metadata.MetaData) (io.WriteCloser, error) {
	// Open a new file in the cloud storage bucket.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create a new cloud storage client")
	}

	timeKey := md.Timestamp().Format("2006/01/02/15/20060102T150405")
	objName := fmt.Sprintf("%s%s_%d.%s", c.prefix, timeKey, md.Seq(), md.Format().Ext())
	if c.Gzip() {
		objName += ".gz"
	}
	obj := client.Bucket(c.bucket).Object(objName)

	w := obj.NewWriter(ctx)
	if c.Gzip() {
		w.ObjectAttrs.ContentEncoding = "gzip"
	}
	return w, nil
}
