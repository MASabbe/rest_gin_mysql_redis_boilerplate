package middleware

import (
	"context"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/tracer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderXRequestID  = "X-Request-ID"
	HeaderXTraceID    = "X-Trace-ID"
	HeaderTraceParent = "traceparent"
)

// RequestID middleware ensures every request has a unique Request ID and distributed Trace ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Resolve or Generate Request ID
		reqID := c.GetHeader(HeaderXRequestID)
		if reqID == "" {
			reqID = uuid.New().String()
		}

		// 2. Resolve or Generate W3C Trace Context / Trace ID
		var tc tracer.TraceContext
		var hasTrace bool

		if traceparent := c.GetHeader(HeaderTraceParent); traceparent != "" {
			tc, hasTrace = tracer.ParseTraceParent(traceparent)
		}
		if !hasTrace {
			if xTraceID := c.GetHeader(HeaderXTraceID); xTraceID != "" {
				tc = tracer.TraceContext{
					TraceID: xTraceID,
					SpanID:  tracer.GenerateSpanID(),
					Sampled: true,
				}
				hasTrace = true
			}
		}
		if !hasTrace {
			tc = tracer.TraceContext{
				TraceID: tracer.GenerateTraceID(),
				SpanID:  tracer.GenerateSpanID(),
				Sampled: true,
			}
		}

		// 3. Set Response Headers
		c.Header(HeaderXRequestID, reqID)
		c.Header(HeaderXTraceID, tc.TraceID)
		c.Header(HeaderTraceParent, tracer.FormatTraceParent(tc))

		// 4. Populate Gin context values
		c.Set("request_id", reqID)
		c.Set("trace_id", tc.TraceID)

		// 5. Populate Go Context
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, logger.RequestIDKey, reqID)
		ctx = tracer.WithTraceContext(ctx, tc)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
