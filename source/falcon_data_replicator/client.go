package falcon_data_replicator

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/m-mizutani/goerr"
	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/pkg/interfaces"
	"github.com/secmon-as-code/hatchery/pkg/logging"
	"github.com/secmon-as-code/hatchery/pkg/metadata"
	"github.com/secmon-as-code/hatchery/pkg/safe"
	"github.com/secmon-as-code/hatchery/pkg/types/secret"
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

type awsConfig struct {
	Region string
	Cred   aws.CredentialsProvider
	SQSURL string
}

type client struct {
	AWS awsConfig

	NewSQS func(cfg aws.Config, optFns ...func(*sqs.Options)) interfaces.SQS
	NewS3  func(cfg aws.Config, optFns ...func(*s3.Options)) interfaces.S3

	MaxMessages int32
	MaxPull     int
}

type Option func(*client)

func WithMaxMessages(n int32) Option {
	return func(x *client) {
		x.MaxMessages = n
	}
}

func WithMaxPull(n int) Option {
	return func(x *client) {
		x.MaxPull = n
	}
}

func WithAWSCredential(cred aws.CredentialsProvider) Option {
	return func(x *client) {
		x.AWS.Cred = cred
	}
}

func New(awsRegion, awsAccessKeyId string, awsSecretAccessKey secret.String, sqsURL string, opts ...Option) hatchery.Source {
	x := &client{
		AWS: awsConfig{
			Region: awsRegion,
			SQSURL: sqsURL,
		},

		NewSQS: func(cfg aws.Config, optFns ...func(*sqs.Options)) interfaces.SQS {
			return sqs.NewFromConfig(cfg, optFns...)
		},
		NewS3: func(cfg aws.Config, optFns ...func(*s3.Options)) interfaces.S3 {
			return s3.NewFromConfig(cfg, optFns...)
		},

		MaxMessages: 10,
		MaxPull:     0,
	}

	for _, opt := range opts {
		opt(x)
	}

	awsOpts := []func(*config.LoadOptions) error{
		config.WithRegion(x.AWS.Region),
	}

	if x.AWS.Cred != nil {
		awsOpts = append(awsOpts,
			config.WithCredentialsProvider(x.AWS.Cred),
		)
	}
	return func(ctx context.Context, p *hatchery.Pipe) error {
		logging.FromCtx(ctx).Info("Run Falcon Data Replicator source", "config", x)

		cfg, err := config.LoadDefaultConfig(ctx, awsOpts...)
		if err != nil {
			return goerr.Wrap(err, "failed to create AWS session")
		}

		// Create AWS service clients
		s3Client := x.NewS3(cfg)
		sqsClient := x.NewSQS(cfg)

		// Receive messages from SQS queue
		input := &sqs.ReceiveMessageInput{
			QueueUrl: aws.String(x.AWS.SQSURL),
		}
		if x.MaxMessages > 0 {
			input.MaxNumberOfMessages = x.MaxMessages
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
	result, err := clients.sqs.ReceiveMessage(ctx, input)
	if err != nil {
		return goerr.Wrap(err, "failed to receive messages from SQS").With("input", input)
	}
	if len(result.Messages) == 0 {
		return errNoMoreMessage
	}

	// Iterate over received messages
	for _, message := range result.Messages {
		// Get the S3 object key from the message
		var msg fdrMessage
		if err := json.Unmarshal([]byte(*message.Body), &msg); err != nil {
			return goerr.Wrap(err, "failed to unmarshal message").With("message", *message.Body)
		}

		for _, file := range msg.Files {
			// Download the object from S3
			s3Input := &s3.GetObjectInput{
				Bucket: aws.String(msg.Bucket),
				Key:    aws.String(file.Path),
			}
			s3Obj, err := clients.s3.GetObject(ctx, s3Input)
			if err != nil {
				return goerr.Wrap(err, "failed to download object from S3").With("msg", msg)
			}
			defer safe.CloseReader(ctx, s3Obj.Body)

			md := metadata.New(
				metadata.WithTimestamp(time.Unix(msg.Timestamp, 0)),
			)
			if err := p.Spout(ctx, s3Obj.Body, md); err != nil {
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
