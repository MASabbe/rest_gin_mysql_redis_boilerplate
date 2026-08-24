package tracer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
)

type spanContextKey struct{}

type TraceContext struct {
	TraceID string
	SpanID  string
	Sampled bool
}

// GenerateTraceID generates a W3C-compliant 128-bit (32 hex characters) trace ID.
func GenerateTraceID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateSpanID generates a W3C-compliant 64-bit (16 hex characters) span ID.
func GenerateSpanID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ParseTraceParent parses a W3C traceparent header: "00-{trace_id}-{parent_id}-{flags}"
func ParseTraceParent(header string) (TraceContext, bool) {
	parts := strings.Split(strings.TrimSpace(header), "-")
	if len(parts) != 4 || parts[0] != "00" {
		return TraceContext{}, false
	}

	traceID := parts[1]
	spanID := parts[2]
	flags := parts[3]

	if len(traceID) != 32 || len(spanID) != 16 || len(flags) != 2 {
		return TraceContext{}, false
	}

	return TraceContext{
		TraceID: traceID,
		SpanID:  spanID,
		Sampled: flags == "01",
	}, true
}

// FormatTraceParent formats a TraceContext into W3C traceparent string.
func FormatTraceParent(tc TraceContext) string {
	flag := "00"
	if tc.Sampled {
		flag = "01"
	}
	return fmt.Sprintf("00-%s-%s-%s", tc.TraceID, tc.SpanID, flag)
}

// WithTraceContext injects TraceContext into Go context.Context and syncs with logger.TraceIDKey.
func WithTraceContext(ctx context.Context, tc TraceContext) context.Context {
	ctx = context.WithValue(ctx, spanContextKey{}, tc)
	ctx = context.WithValue(ctx, logger.TraceIDKey, tc.TraceID)
	return ctx
}

// TraceFromContext extracts TraceContext from Go context.Context.
func TraceFromContext(ctx context.Context) (TraceContext, bool) {
	if ctx == nil {
		return TraceContext{}, false
	}
	tc, ok := ctx.Value(spanContextKey{}).(TraceContext)
	if ok && tc.TraceID != "" {
		return tc, true
	}
	// Fallback to logger.TraceIDKey if present
	if traceID, ok := ctx.Value(logger.TraceIDKey).(string); ok && traceID != "" {
		return TraceContext{TraceID: traceID}, true
	}
	return TraceContext{}, false
}

// TraceIDFromContext returns the trace ID string or empty string.
func TraceIDFromContext(ctx context.Context) string {
	tc, ok := TraceFromContext(ctx)
	if !ok {
		return ""
	}
	return tc.TraceID
}

// Span is an OpenTelemetry-compatible span abstraction.
type Span interface {
	End()
	SetAttribute(key string, val any)
}

type noopSpan struct{}

func (s *noopSpan) End()                             {}
func (s *noopSpan) SetAttribute(key string, val any) {}

// Tracer is an OpenTelemetry-compatible tracer abstraction.
type Tracer interface {
	Start(ctx context.Context, spanName string) (context.Context, Span)
}

type NoopTracer struct{}

func (t *NoopTracer) Start(ctx context.Context, spanName string) (context.Context, Span) {
	tc, ok := TraceFromContext(ctx)
	if !ok {
		tc = TraceContext{
			TraceID: GenerateTraceID(),
			SpanID:  GenerateSpanID(),
			Sampled: true,
		}
		ctx = WithTraceContext(ctx, tc)
	}
	return ctx, &noopSpan{}
}
