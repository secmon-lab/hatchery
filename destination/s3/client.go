package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/m-mizutani/goerr"
	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/pkg/logging"
	"github.com/secmon-as-code/hatchery/pkg/metadata"
)

type client struct {
	region      string
	bucket      string
	prefix      string
	cred        aws.CredentialsProvider
	objNameFunc ObjNameFunc
}

type Option func(*client)

func WithPrefix(prefix string) Option {
	return func(c *client) {
		c.prefix = prefix
	}
}

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

type pipeWrier struct {
	w     io.WriteCloser
	errCh chan error
}

func (x *pipeWrier) Write(p []byte) (n int, err error) {
	return x.w.Write(p)
}

func (x *pipeWrier) Close() error {
	if err := x.w.Close(); err != nil {
		return goerr.Wrap(err, "failed to close write buffer")
	}

	if err := <-x.errCh; err != nil {
		return goerr.Wrap(err, "failed to write buffer")
	}

	return nil
}

func New(region, bucket string, options ...Option) hatchery.Destination {
	client := &client{
		bucket:      bucket,
		region:      region,
		objNameFunc: DefaultObjectName,
	}

	for _, opt := range options {
		opt(client)
	}

	awsOpts := []func(*config.LoadOptions) error{
		config.WithRegion(client.region),
	}

	if client.cred != nil {
		awsOpts = append(awsOpts,
			config.WithCredentialsProvider(client.cred),
		)
	}

	return func(ctx context.Context, md metadata.MetaData) (io.WriteCloser, error) {
		cfg, err := config.LoadDefaultConfig(ctx, awsOpts...)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to create AWS session")
		}

		// Create AWS service clients
		s3Client := s3.NewFromConfig(cfg)

		args := ObjNameArgs{
			Prefix:    client.prefix,
			Timestamp: md.Timestamp(),
			Seq:       md.Seq(),
			Ext:       md.Format().Ext(),
		}
		objName := client.objNameFunc(args)

		errCh := make(chan error, 1)
		r, w := io.Pipe()
		pipe := &pipeWrier{
			w:     w,
			errCh: errCh,
		}

		go func() {
			defer close(errCh)

			uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
				u.PartSize = 10 * 1024 * 1024
			})

			input := &s3.PutObjectInput{
				Bucket: aws.String(client.bucket),
				Key:    aws.String(objName),
				Body:   r,
			}

			logging.FromCtx(ctx).Info("Start to put object", "bucket", client.bucket, "key", objName)
			if _, err := uploader.Upload(ctx, input); err != nil {
				errCh <- goerr.Wrap(err, "failed to put object")
				return
			}

			if err := r.Close(); err != nil {
				errCh <- goerr.Wrap(err, "failed to close read buffer")
				return
			}
		}()

		return pipe, nil
	}
}
