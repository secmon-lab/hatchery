package falcon_data_replicator

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
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
	Cred   *credentials.Credentials
	SQSURL string
}

type config struct {
	AWS awsConfig

	NewSQS func(*session.Session) interfaces.SQS
	NewS3  func(*session.Session) interfaces.S3

	MaxMessages int64
	MaxPull     int
}

type Option func(*config)

func WithMaxMessages(n int64) Option {
	return func(x *config) {
		x.MaxMessages = n
	}
}

func WithMaxPull(n int) Option {
	return func(x *config) {
		x.MaxPull = n
	}
}

func WithAWSCredential(cred *credentials.Credentials) Option {
	return func(x *config) {
		x.AWS.Cred = cred
	}
}

// WithSQSClientFactory sets a factory function to create an SQS client. This option is mainly for testing.
func WithSQSClientFactory(f func(*session.Session) interfaces.SQS) Option {
	return func(x *config) {
		x.NewSQS = f
	}
}

// WithS3ClientFactory sets a factory function to create an S3 client. This option is mainly for testing.
func WithS3ClientFactory(f func(*session.Session) interfaces.S3) Option {
	return func(x *config) {
		x.NewS3 = f
	}
}

func New(awsRegion, awsAccessKeyId string, awsSecretAccessKey secret.String, sqsURL string, opts ...Option) hatchery.Source {
	x := &config{
		AWS: awsConfig{
			Region: awsRegion,
			SQSURL: sqsURL,
		},

		NewSQS: func(ssn *session.Session) interfaces.SQS {
			return sqs.New(ssn)
		},
		NewS3: func(ssn *session.Session) interfaces.S3 {
			return s3.New(ssn)
		},

		MaxMessages: 10,
		MaxPull:     0,
	}

	for _, opt := range opts {
		opt(x)
	}

	return func(ctx context.Context, p *hatchery.Pipe) error {
		logging.FromCtx(ctx).Info("Run Falcon Data Replicator source", "config", x)

		// Create an AWS session
		awsSession, err := session.NewSession(&aws.Config{
			Region:      aws.String(x.AWS.Region),
			Credentials: x.AWS.Cred,
		})
		if err != nil {
			return goerr.Wrap(err, "failed to create AWS session").With("client", x)
		}

		// Create AWS service clients
		sqsClient := x.NewSQS(awsSession)
		s3Client := x.NewS3(awsSession)

		// Receive messages from SQS queue
		input := &sqs.ReceiveMessageInput{
			QueueUrl: aws.String(x.AWS.SQSURL),
		}
		if x.MaxMessages > 0 {
			input.MaxNumberOfMessages = aws.Int64(x.MaxMessages)
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
	result, err := clients.sqs.ReceiveMessageWithContext(ctx, input)
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
			s3Obj, err := clients.s3.GetObjectWithContext(ctx, s3Input)
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
		_, err = clients.sqs.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      input.QueueUrl,
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			return goerr.Wrap(err, "failed to delete message from SQS")
		}
	}

	return nil
}
