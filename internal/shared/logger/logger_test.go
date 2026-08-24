package logger_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogger_Init_And_Sanitization(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logger.Init("production", "info", buf)
	assert.NotNil(t, l)

	ctx := context.WithValue(context.Background(), logger.RequestIDKey, "req-12345")
	ctx = context.WithValue(ctx, logger.UserIDKey, "usr-999")

	logger.WithContext(ctx).Info("user login attempt",
		slog.String("email", "test@example.com"),
		slog.String("password", "supersecret123"),
		slog.String("token", "jwt-token-value"),
	)

	output := buf.String()
	assert.Contains(t, output, "req-12345")
	assert.Contains(t, output, "usr-999")
	assert.Contains(t, output, "test@example.com")
	assert.Contains(t, output, `"[REDACTED]"`)
	assert.NotContains(t, output, "supersecret123")
	assert.NotContains(t, output, "jwt-token-value")
}
