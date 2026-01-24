package awsclient_test

import (
	"context"
	"io"
	"strings"
	"testing"

	awsclient "github.com/bpurdy1/aws-client"
	"github.com/bpurdy1/aws-client/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMockClient_PutObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockClient(ctrl)
	ctx := context.Background()

	mockClient.EXPECT().
		PutObject(ctx, "test-bucket", "test-key", gomock.Any()).
		Return(nil)

	err := mockClient.PutObject(ctx, "test-bucket", "test-key", strings.NewReader("test content"))
	assert.NoError(t, err)
}

func TestMockClient_GetObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockClient(ctrl)
	ctx := context.Background()

	expectedContent := "test content"
	mockClient.EXPECT().
		GetObject(ctx, "test-bucket", "test-key").
		Return(io.NopCloser(strings.NewReader(expectedContent)), nil)

	reader, err := mockClient.GetObject(ctx, "test-bucket", "test-key")
	assert.NoError(t, err)

	content, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, string(content))
}

func TestMockClient_DeleteObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockClient(ctrl)
	ctx := context.Background()

	mockClient.EXPECT().
		DeleteObject(ctx, "test-bucket", "test-key").
		Return(nil)

	err := mockClient.DeleteObject(ctx, "test-bucket", "test-key")
	assert.NoError(t, err)
}

func TestMockClient_SendMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockClient(ctrl)
	ctx := context.Background()

	expectedMessageID := "msg-123"
	mockClient.EXPECT().
		SendMessage(ctx, "https://sqs.us-east-1.amazonaws.com/123456789/test-queue", "test message").
		Return(expectedMessageID, nil)

	messageID, err := mockClient.SendMessage(ctx, "https://sqs.us-east-1.amazonaws.com/123456789/test-queue", "test message")
	assert.NoError(t, err)
	assert.Equal(t, expectedMessageID, messageID)
}

func TestMockClient_ReceiveMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockClient(ctrl)
	ctx := context.Background()

	expectedMessages := []awsclient.Message{
		{ID: "msg-1", Body: "message 1", ReceiptHandle: "handle-1"},
		{ID: "msg-2", Body: "message 2", ReceiptHandle: "handle-2"},
	}

	mockClient.EXPECT().
		ReceiveMessages(ctx, "https://sqs.us-east-1.amazonaws.com/123456789/test-queue", int32(10)).
		Return(expectedMessages, nil)

	messages, err := mockClient.ReceiveMessages(ctx, "https://sqs.us-east-1.amazonaws.com/123456789/test-queue", 10)
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, expectedMessages, messages)
}

func TestMockClient_DeleteMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockClient(ctrl)
	ctx := context.Background()

	mockClient.EXPECT().
		DeleteMessage(ctx, "https://sqs.us-east-1.amazonaws.com/123456789/test-queue", "receipt-handle-123").
		Return(nil)

	err := mockClient.DeleteMessage(ctx, "https://sqs.us-east-1.amazonaws.com/123456789/test-queue", "receipt-handle-123")
	assert.NoError(t, err)
}

func TestLoadConfig(t *testing.T) {
	t.Setenv("AWS_REGION", "us-west-2")
	t.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")

	cfg, err := awsclient.LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "us-west-2", cfg.Region)
	assert.Equal(t, "test-key", cfg.AccessKeyID)
	assert.Equal(t, "test-secret", cfg.SecretAccessKey)
}

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")

	cfg, err := awsclient.LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", cfg.Region)
}
