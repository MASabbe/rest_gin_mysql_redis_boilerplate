package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	TraceIDKey   contextKey = "trace_id"
	UserIDKey    contextKey = "user_id"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(defaultLogger)
}

// Init configures the global logger based on environment and log level.
func Init(env, level string, out ...io.Writer) *slog.Logger {
	var writer io.Writer = os.Stdout
	if len(out) > 0 && out[0] != nil {
		writer = out[0]
	}

	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:       lvl,
		ReplaceAttr: sanitizeAttr,
	}

	var handler slog.Handler
	if strings.ToLower(env) == "production" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
	return defaultLogger
}

// Get returns the default structured logger.
func Get() *slog.Logger {
	return defaultLogger
}

// WithContext returns a contextual logger populated with RequestID, TraceID, and UserID if present in ctx.
func WithContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return defaultLogger
	}

	var attrs []any
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		attrs = append(attrs, slog.String("request_id", reqID))
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		attrs = append(attrs, slog.String("trace_id", traceID))
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		attrs = append(attrs, slog.String("user_id", userID))
	}

	if len(attrs) == 0 {
		return defaultLogger
	}

	return defaultLogger.With(attrs...)
}

// sanitizeAttr redacts sensitive fields like passwords and tokens.
func sanitizeAttr(groups []string, a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)
	sensitiveKeys := []string{
		"password", "password_hash", "secret", "token", "access_token",
		"refresh_token", "jwt", "authorization", "db_password", "apikey",
	}

	for _, sk := range sensitiveKeys {
		if strings.Contains(key, sk) {
			return slog.String(a.Key, "[REDACTED]")
		}
	}
	return a
}
