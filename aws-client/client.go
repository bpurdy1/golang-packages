//go:generate mockgen -source=client.go -destination=./mock/mock.go -package=mock

package awsclient

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Client interface {
	// S3 operations
	PutObject(ctx context.Context, bucket, key string, body io.Reader) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucket, key string) error

	// SQS operations
	SendMessage(ctx context.Context, queueURL, messageBody string) (string, error)
	ReceiveMessages(ctx context.Context, queueURL string, maxMessages int32) ([]Message, error)
	DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error
}

// Message represents an SQS message.
type Message struct {
	ID            string
	Body          string
	ReceiptHandle string
}

type AWSClient struct {
	s3Client  *s3.Client
	sqsClient *sqs.Client
	cfg       *Config
}

func New(ctx context.Context, cfg *Config) (*AWSClient, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				cfg.SessionToken,
			),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	s3Opts := []func(*s3.Options){}
	sqsOpts := []func(*sqs.Options){}

	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		})
		sqsOpts = append(sqsOpts, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	return &AWSClient{
		s3Client:  s3.NewFromConfig(awsCfg, s3Opts...),
		sqsClient: sqs.NewFromConfig(awsCfg, sqsOpts...),
		cfg:       cfg,
	}, nil
}

// PutObject uploads an object to S3.
func (c *AWSClient) PutObject(ctx context.Context, bucket, key string, body io.Reader) error {
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

// GetObject retrieves an object from S3.
func (c *AWSClient) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	output, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

// DeleteObject removes an object from S3.
func (c *AWSClient) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// SendMessage sends a message to an SQS queue.
func (c *AWSClient) SendMessage(ctx context.Context, queueURL, messageBody string) (string, error) {
	output, err := c.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(messageBody),
	})
	if err != nil {
		return "", err
	}
	return *output.MessageId, nil
}

// ReceiveMessages receives messages from an SQS queue.
func (c *AWSClient) ReceiveMessages(ctx context.Context, queueURL string, maxMessages int32) ([]Message, error) {
	output, err := c.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: maxMessages,
	})
	if err != nil {
		return nil, err
	}

	messages := make([]Message, len(output.Messages))
	for i, msg := range output.Messages {
		messages[i] = Message{
			ID:            *msg.MessageId,
			Body:          *msg.Body,
			ReceiptHandle: *msg.ReceiptHandle,
		}
	}
	return messages, nil
}

// DeleteMessage deletes a message from an SQS queue.
func (c *AWSClient) DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error {
	_, err := c.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}
