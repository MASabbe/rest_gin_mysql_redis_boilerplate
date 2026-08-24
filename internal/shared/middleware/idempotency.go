package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	sharedResponse "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type idempotencyStatus string

const (
	statusProcessing idempotencyStatus = "PROCESSING"
	statusCompleted  idempotencyStatus = "COMPLETED"
)

type IdempotencyRecord struct {
	Fingerprint string            `json:"fingerprint"`
	Status      idempotencyStatus `json:"status"`
	StatusCode  int               `json:"status_code"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Idempotency provides idempotency key handling backed by Redis.
// Applied to state-mutating requests (POST, PUT, PATCH, DELETE) when 'Idempotency-Key' header is supplied.
func Idempotency(redisClient *redis.Client, ttl time.Duration) gin.HandlerFunc {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
		if key == "" {
			c.Next()
			return
		}

		// Only apply to mutating methods
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
			c.Next()
			return
		}

		if redisClient == nil {
			c.Next()
			return
		}

		// Read and preserve request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Calculate payload fingerprint
		hasher := sha256.New()
		hasher.Write([]byte(method))
		hasher.Write([]byte(c.Request.URL.Path))
		hasher.Write(bodyBytes)
		fingerprint := hex.EncodeToString(hasher.Sum(nil))

		cacheKey := "idempotency:" + key
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()

		// 1. Check existing record
		existingJSON, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil && existingJSON != "" {
			var record IdempotencyRecord
			if jsonErr := json.Unmarshal([]byte(existingJSON), &record); jsonErr == nil {
				if record.Fingerprint != fingerprint {
					sharedResponse.Error(c, appErrors.NewBusinessError(
						"IDEMPOTENCY_CONFLICT",
						"Idempotency-Key was already used with a different request payload",
					))
					c.Abort()
					return
				}

				if record.Status == statusProcessing {
					sharedResponse.Error(c, appErrors.NewConflictError(
						"A request with this Idempotency-Key is currently in progress",
					))
					c.Abort()
					return
				}

				if record.Status == statusCompleted {
					c.Header("X-Idempotent-Replay", "true")
					for k, v := range record.Headers {
						c.Header(k, v)
					}
					c.Data(record.StatusCode, "application/json; charset=utf-8", []byte(record.Body))
					c.Abort()
					return
				}
			}
		}

		// 2. Lock key with PROCESSING status
		initialRecord := IdempotencyRecord{
			Fingerprint: fingerprint,
			Status:      statusProcessing,
		}
		initBytes, _ := json.Marshal(initialRecord)
		acquired, err := redisClient.SetNX(ctx, cacheKey, string(initBytes), 30*time.Second).Result()
		if err != nil {
			logger.Get().Warn("Failed to acquire idempotency lock in Redis, proceeding", "error", err)
			c.Next()
			return
		}
		if !acquired {
			sharedResponse.Error(c, appErrors.NewConflictError(
				"A request with this Idempotency-Key is currently in progress",
			))
			c.Abort()
			return
		}

		// 3. Capture downstream execution
		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = writer

		c.Next()

		// 4. Save response or release on server error
		statusCode := c.Writer.Status()
		if statusCode >= 500 {
			_ = redisClient.Del(context.Background(), cacheKey).Err()
			return
		}

		completedRecord := IdempotencyRecord{
			Fingerprint: fingerprint,
			Status:      statusCompleted,
			StatusCode:  statusCode,
			Headers: map[string]string{
				"Content-Type": c.Writer.Header().Get("Content-Type"),
			},
			Body: writer.body.String(),
		}
		compBytes, _ := json.Marshal(completedRecord)
		_ = redisClient.Set(context.Background(), cacheKey, string(compBytes), ttl).Err()
	}
}
