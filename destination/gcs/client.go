package gcs

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/m-mizutani/goerr"
	"github.com/secmon-lab/hatchery"
	"github.com/secmon-lab/hatchery/pkg/logging"
	"github.com/secmon-lab/hatchery/pkg/metadata"
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
	Prefix     string
	Timestamp  time.Time
	Seq        int
	Ext        string
	SchemaHint string
	Slug       string
}

type ObjNameFunc func(args ObjNameArgs) string

func DefaultObjectName(args ObjNameArgs) string {
	timeKey := args.Timestamp.Format("2006/01/02/15/20060102T150405")
	schema := args.SchemaHint
	if schema != "" {
		schema += "/"
	}

	var slug string
	if args.Slug != "" {
		slug = "_" + args.Slug
	}
	return fmt.Sprintf("%s%s%s%s_%04d.%s", args.Prefix, schema, timeKey, slug, args.Seq, args.Ext)
}

type gzipWriter struct {
	writer     io.WriteCloser
	gzipWriter *gzip.Writer
}

func (w *gzipWriter) Write(p []byte) (n int, err error) {
	return w.gzipWriter.Write(p)
}

func (w *gzipWriter) Close() error {
	if err := w.gzipWriter.Close(); err != nil {
		return goerr.Wrap(err, "failed to close gzip writer")
	}
	if err := w.writer.Close(); err != nil {
		return goerr.Wrap(err, "failed to close writer")
	}
	return nil
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
			Prefix:     c.prefix,
			Timestamp:  md.Timestamp(),
			Seq:        md.Seq(),
			Ext:        md.Format().Ext(),
			SchemaHint: md.SchemaHint(),
			Slug:       md.Slug(),
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
			w = &gzipWriter{
				writer:     objWriter,
				gzipWriter: gzip.NewWriter(objWriter),
			}
		}

		logging.FromCtx(ctx).Info("New destination (Google Cloud Storage)", "bucket", c.bucket, "object", objName, "metadata", md)

		return w, nil
	}
}

type Option func(*Client)

// WithPrefix sets a prefix for object names in the bucket.
func WithPrefix(prefix string) Option {
	return func(c *Client) {
		c.prefix = prefix
	}
}

// WithGzip sets a flag to compress data with gzip.
func WithGzip(gzip bool) Option {
	return func(c *Client) {
		c.gzip = gzip
	}
}

func WithClientOptions(options ...option.ClientOption) Option {
	return func(c *Client) {
		c.options = append(c.options, options...)
	}
}
