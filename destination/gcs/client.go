package gcs

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/m-mizutani/goerr"
	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/pkg/metadata"
	"google.golang.org/api/option"
)

// Client is a destination that writes data to a Google cloud storage bucket.
type Client struct {
	bucket      string
	prefix      string
	gzip        bool
	objNameFunc ObjNameFunc
	options     []option.ClientOption
}

func (c *Client) Bucket() string { return c.bucket }
func (c *Client) Prefix() string { return c.prefix }
func (c *Client) Gzip() bool     { return c.gzip }

type ObjNameArgs struct {
	Prefix    string
	Timestamp time.Time
	Seq       int
	Ext       string
}

type ObjNameFunc func(args ObjNameArgs) string

func DefaultObjectName(args ObjNameArgs) string {
	timeKey := args.Timestamp.Format("2006/01/02/15/20060102T150405")
	return fmt.Sprintf("%s%s_%04d.%s", args.Prefix, timeKey, args.Seq, args.Ext)
}

// New creates a new Client destination.
func New(bucket string, options ...Option) hatchery.Destination {
	c := &Client{
		bucket:      bucket,
		objNameFunc: DefaultObjectName,
	}

	for _, opt := range options {
		opt(c)
	}

	return func(ctx context.Context, md metadata.MetaData) (io.WriteCloser, error) {
		// Open a new file in the cloud storage bucket.
		client, err := storage.NewClient(ctx, c.options...)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to create a new cloud storage client")
		}

		args := ObjNameArgs{
			Prefix:    c.prefix,
			Timestamp: md.Timestamp(),
			Seq:       md.Seq(),
			Ext:       md.Format().Ext(),
		}
		if c.gzip {
			args.Ext += ".gz"
		}

		objName := c.objNameFunc(args)

		obj := client.Bucket(c.bucket).Object(objName)
		objWriter := obj.NewWriter(ctx)
		var w io.WriteCloser = objWriter
		if c.gzip {
			objWriter.ObjectAttrs.ContentEncoding = "gzip"
			w = gzip.NewWriter(objWriter)
		}
		return w, nil
	}
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

func WithClientOptions(options ...option.ClientOption) Option {
	return func(c *Client) error {
		c.options = append(c.options, options...)
		return nil
	}
}