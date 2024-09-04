package falcon_data_replicator

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/m-mizutani/goerr"
	"github.com/secmon-as-code/hatchery"
	"github.com/secmon-as-code/hatchery/pkg/interfaces"
	"github.com/secmon-as-code/hatchery/pkg/types"
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

type Client struct {
	awsRegion          string
	awsAccessKeyId     string
	awsSecretAccessKey types.SecretString
	sqsURL             string

	maxMessages int64
	maxPull     int

	sqs interfaces.SQS
	s3  interfaces.S3
}

type Option func(*Client)

func WithMaxMessages(n int64) Option {
	return func(x *Client) {
		x.maxMessages = n
	}
}

func WithMaxPull(n int) Option {
	return func(x *Client) {
		x.maxPull = n
	}
}

func New(awsRegion, awsAccessKeyId string, awsSecretAccessKey types.SecretString, sqsURL string, opts ...Option) *Client {
	x := &Client{
		awsRegion:          awsRegion,
		awsAccessKeyId:     awsAccessKeyId,
		awsSecretAccessKey: awsSecretAccessKey,
		sqsURL:             sqsURL,

		maxMessages: 10,
		maxPull:     0,
	}

	for _, opt := range opts {
		opt(x)
	}

	return x
}

func (x *Client) Load(ctx context.Context, p *hatchery.Pipe) error {
	// Create an AWS session
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(x.awsRegion),
		Credentials: credentials.NewCredentials(&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     x.awsAccessKeyId,
				SecretAccessKey: x.awsSecretAccessKey.UnsafeString(),
			},
		}),
	})
	if err != nil {
		return goerr.Wrap(err, "failed to create AWS session").With("client", x)
	}

	// Create AWS service clients
	sqsClient := clients.NewSQS(awsSession)
	s3Client := clients.NewS3(awsSession)

	// Receive messages from SQS queue
	input := &sqs.ReceiveMessageInput{
		QueueUrl: aws.String(x.sqsURL),
	}
	if x.maxMessages > 0 {
		input.MaxNumberOfMessages = aws.Int64(x.maxMessages)
	}

	for i := 0; x.maxPull == 0 || i < x.maxPull; i++ {
		c := &fdrClients{infra: clients, sqs: sqsClient, s3: s3Client}
		if err := copy(ctx, c, input, types.CSBucket(req.Bucket), prefix); err != nil {
			if err == errNoMoreMessage {
				break
			}
			return err
		}
	}

	return nil
}

var (
	errNoMoreMessage = errors.New("no more message")
)

func copy(ctx context.Context, clients *fdrClients, input *sqs.ReceiveMessageInput, bucket types.CSBucket, prefix types.CSObjectName) error {
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
			defer utils.SafeClose(s3Obj.Body)

			csObj := prefix + types.CSObjectName(file.Path)
			w := clients.infra.CloudStorage().NewObjectWriter(ctx, bucket, csObj)

			if _, err := io.Copy(w, s3Obj.Body); err != nil {
				return goerr.Wrap(err, "failed to write object to GCS").With("msg", msg)
			}
			if err := w.Close(); err != nil {
				return goerr.Wrap(err, "failed to close object writer").With("msg", msg)
			}

			utils.CtxLogger(ctx).Info("FDR: object forwarded from S3 to GCS", "s3", s3Input, "gcsObj", csObj)
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
