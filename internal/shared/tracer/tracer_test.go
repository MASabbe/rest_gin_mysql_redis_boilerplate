package tracer_test

import (
	"context"
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/tracer"
	"github.com/stretchr/testify/assert"
)

func TestTracer_IDGeneration(t *testing.T) {
	traceID := tracer.GenerateTraceID()
	assert.Len(t, traceID, 32)

	spanID := tracer.GenerateSpanID()
	assert.Len(t, spanID, 16)
}

func TestTracer_W3CTraceParent(t *testing.T) {
	t.Run("Parse and format valid header", func(t *testing.T) {
		raw := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
		tc, ok := tracer.ParseTraceParent(raw)
		assert.True(t, ok)
		assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", tc.TraceID)
		assert.Equal(t, "00f067aa0ba902b7", tc.SpanID)
		assert.True(t, tc.Sampled)

		formatted := tracer.FormatTraceParent(tc)
		assert.Equal(t, raw, formatted)
	})

	t.Run("Reject malformed header", func(t *testing.T) {
		_, ok := tracer.ParseTraceParent("invalid-trace-parent")
		assert.False(t, ok)

		_, ok = tracer.ParseTraceParent("01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01") // wrong version
		assert.False(t, ok)
	})
}

func TestTracer_ContextPropagation(t *testing.T) {
	ctx := context.Background()
	assert.Empty(t, tracer.TraceIDFromContext(ctx))

	tc := tracer.TraceContext{
		TraceID: "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:  "00f067aa0ba902b7",
		Sampled: true,
	}

	ctx = tracer.WithTraceContext(ctx, tc)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", tracer.TraceIDFromContext(ctx))
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", ctx.Value(logger.TraceIDKey))

	extracted, ok := tracer.TraceFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, tc, extracted)
}

func TestNoopTracer(t *testing.T) {
	tr := &tracer.NoopTracer{}
	ctx, span := tr.Start(context.Background(), "test.operation")
	defer span.End()

	span.SetAttribute("key", "val")
	assert.NotEmpty(t, tracer.TraceIDFromContext(ctx))
}
