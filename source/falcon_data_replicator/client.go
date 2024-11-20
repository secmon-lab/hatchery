package falcon_data_replicator

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/m-mizutani/goerr"
	"github.com/secmon-lab/hatchery"
	"github.com/secmon-lab/hatchery/pkg/interfaces"
	"github.com/secmon-lab/hatchery/pkg/logging"
	"github.com/secmon-lab/hatchery/pkg/metadata"
	"github.com/secmon-lab/hatchery/pkg/safe"
	"github.com/secmon-lab/hatchery/pkg/types/secret"
)

type fdrMessage struct {
	Bucket     string `json:"bucket"`
	Cid        string `json:"cid"`
	FileCount  int64  `json:"fileCount"`
	Files      []file `json:"files"`
	PathPrefix string `json:"pathPrefix"`
	Timestamp  int64  `json:"timestamp"`
	TotalSize  int64  `json:"totalSize"`
}

type file struct {
	Checksum string `json:"checksum"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
}

// Export the config struct for logging
type awsConfig struct {
	Region string
	cred   aws.CredentialsProvider
	SqsURL string
}

type client struct {
	AWS awsConfig

	NewSQS func(cfg aws.Config, optFns ...func(*sqs.Options)) interfaces.SQS `json:"-"`
	NewS3  func(cfg aws.Config, optFns ...func(*s3.Options)) interfaces.S3   `json:"-"`

	MaxPull int
}

type Option func(*client)

func WithMaxPull(n int) Option {
	return func(x *client) {
		x.MaxPull = n
	}
}

func WithAWSCredential(cred aws.CredentialsProvider) Option {
	return func(x *client) {
		x.AWS.cred = cred
	}
}

func New(awsRegion, awsAccessKeyId string, awsSecretAccessKey secret.String, sqsURL string, opts ...Option) hatchery.Source {
	x := &client{
		AWS: awsConfig{
			Region: awsRegion,
			SqsURL: sqsURL,
			cred:   credentials.NewStaticCredentialsProvider(awsAccessKeyId, awsSecretAccessKey.Unsafe(), ""),
		},

		NewSQS: func(cfg aws.Config, optFns ...func(*sqs.Options)) interfaces.SQS {
			return sqs.NewFromConfig(cfg, optFns...)
		},
		NewS3: func(cfg aws.Config, optFns ...func(*s3.Options)) interfaces.S3 {
			return s3.NewFromConfig(cfg, optFns...)
		},

		MaxPull: 0,
	}

	for _, opt := range opts {
		opt(x)
	}

	awsOpts := []func(*config.LoadOptions) error{
		config.WithRegion(x.AWS.Region),
	}

	if x.AWS.cred != nil {
		awsOpts = append(awsOpts,
			config.WithCredentialsProvider(x.AWS.cred),
		)
	}
	return func(ctx context.Context, p *hatchery.Pipe) error {
		logger := logging.FromCtx(ctx).With("source", "falcon_data_replicator")
		logger.Info("New source (Falcon Data Replicator)", "config", x)
		ctx = logging.InjectCtx(ctx, logger)

		cfg, err := config.LoadDefaultConfig(ctx, awsOpts...)
		if err != nil {
			return goerr.Wrap(err, "failed to create AWS session")
		}

		// Create AWS service clients
		s3Client := x.NewS3(cfg)
		sqsClient := x.NewSQS(cfg)

		// Receive messages from SQS queue
		input := &sqs.ReceiveMessageInput{
			QueueUrl: aws.String(x.AWS.SqsURL),
		}

		for i := 0; x.MaxPull == 0 || i < x.MaxPull; i++ {
			c := &fdrClients{sqs: sqsClient, s3: s3Client}
			if err := copy(ctx, c, input, p); err != nil {
				if err == errNoMoreMessage {
					break
				}
				return err
			}
		}

		return nil
	}
}

type fdrClients struct {
	sqs interfaces.SQS
	s3  interfaces.S3
}

var (
	errNoMoreMessage = errors.New("no more message")
)

func copy(ctx context.Context, clients *fdrClients, input *sqs.ReceiveMessageInput, p *hatchery.Pipe) error {
	logger := logging.FromCtx(ctx)
	result, err := clients.sqs.ReceiveMessage(ctx, input)
	if err != nil {
		return goerr.Wrap(err, "failed to receive messages from SQS").With("input", input)
	}
	if len(result.Messages) == 0 {
		return errNoMoreMessage
	}

	// Iterate over received messages
	for _, message := range result.Messages {
		if message.Body == nil {
			logger.Warn("Received message with no body", "message", message)
			continue
		}

		// Get the S3 object key from the message
		var msg fdrMessage
		if err := json.Unmarshal([]byte(*message.Body), &msg); err != nil {
			return goerr.Wrap(err, "failed to unmarshal message").With("message", *message.Body)
		}

		logger.Debug("Received SQS message", "msg", msg)

		for _, file := range msg.Files {
			// Download the object from S3
			logger.Info("downloading object from S3", "bucket", msg.Bucket, "path", file.Path)
			s3Input := &s3.GetObjectInput{
				Bucket: aws.String(msg.Bucket),
				Key:    aws.String(file.Path),
			}
			s3Obj, err := clients.s3.GetObject(ctx, s3Input)
			if err != nil {
				return goerr.Wrap(err, "failed to download object from S3").With("msg", msg)
			}
			defer safe.CloseReader(ctx, s3Obj.Body)

			// Parse key of the object
			parts := strings.Split(file.Path, "/")
			var schemaHint string
			if len(parts) > 1 {
				switch parts[1] {
				case "data":
					schemaHint = "data"
				case "fdrv2":
					schemaHint = "fdrv2_" + parts[2]
				}
			}
			if schemaHint == "" {
				logger.Warn("failed to parse schema hint", "path", file.Path)
				schemaHint = "unknown"
			}

			pathHash := sha256.Sum256([]byte(file.Path))
			md := metadata.New(
				metadata.WithTimestamp(time.Unix(msg.Timestamp/1000, 0)),
				metadata.WithSchemaHint(schemaHint),
				metadata.WithSlug(hex.EncodeToString(pathHash[:])),
			)

			r, err := gzip.NewReader(s3Obj.Body)
			if err != nil {
				return goerr.Wrap(err, "failed to create gzip reader").With("msg", msg)
			}

			if err := p.Spout(ctx, r, md); err != nil {
				return goerr.Wrap(err, "failed to write object to destination").With("msg", msg)
			}
		}

		// Delete the message from SQS
		_, err = clients.sqs.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      input.QueueUrl,
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			return goerr.Wrap(err, "failed to delete message from SQS")
		}
	}

	return nil
}
