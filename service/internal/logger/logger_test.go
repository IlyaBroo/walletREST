package logger

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestLogger struct {
	logBuffer *bytes.Buffer
	Logger    Logger
}

func NewTestLogger() *TestLogger {
	logBuffer := new(bytes.Buffer)

	cfg := &Config{
		Path:         "",
		Level:        "local",
		Service_name: "test_service",
		Writer:       logBuffer,
	}

	lg, err := NewLogger(WithCfg(cfg))
	if err != nil {
		panic(err)
	}

	return &TestLogger{
		logBuffer: logBuffer,
		Logger:    lg,
	}
}

func TestLogger_DebugCtx(t *testing.T) {
	testLogger := NewTestLogger()
	ctx := context.WithValue(context.Background(), "requestID", "12345")

	testLogger.Logger.DebugCtx(ctx, "Debug message")

	assert.Contains(t, testLogger.logBuffer.String(), "Debug message")
	assert.Contains(t, testLogger.logBuffer.String(), "request_ID=12345")
	assert.Contains(t, testLogger.logBuffer.String(), "service_name=test_service")
}

func TestLogger_InfoCtx(t *testing.T) {
	testLogger := NewTestLogger()
	ctx := context.WithValue(context.Background(), "requestID", "67890")

	testLogger.Logger.InfoCtx(ctx, "Info message")

	assert.Contains(t, testLogger.logBuffer.String(), "Info message")
	assert.Contains(t, testLogger.logBuffer.String(), "request_ID=67890")
	assert.Contains(t, testLogger.logBuffer.String(), "service_name=test_service")
}

func TestLogger_WarnCtx(t *testing.T) {
	testLogger := NewTestLogger()
	ctx := context.WithValue(context.Background(), "requestID", "54321")

	testLogger.Logger.WarnCtx(ctx, "Warning message")

	assert.Contains(t, testLogger.logBuffer.String(), "Warning message")
	assert.Contains(t, testLogger.logBuffer.String(), "request_ID=54321")
	assert.Contains(t, testLogger.logBuffer.String(), "service_name=test_service")
}

func TestLogger_ErrorCtx(t *testing.T) {
	testLogger := NewTestLogger()
	ctx := context.WithValue(context.Background(), "requestID", "98765")

	testLogger.Logger.ErrorCtx(ctx, "Error message")

	assert.Contains(t, testLogger.logBuffer.String(), "Error message")
	assert.Contains(t, testLogger.logBuffer.String(), "request_ID=98765")
	assert.Contains(t, testLogger.logBuffer.String(), "service_name=test_service")
}
