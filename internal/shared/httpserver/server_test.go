package httpserver_test

import (
	"context"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/httpserver"
	"github.com/stretchr/testify/assert"
)

func TestServer_New_And_RegisterCleanup(t *testing.T) {
	appCfg := config.AppConfig{
		Name: "test-app",
		Env:  "testing",
	}
	srvCfg := config.ServerConfig{
		Host:              "127.0.0.1",
		Port:              "9999",
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ShutdownTimeout:   5 * time.Second,
		MaxBodySizeBytes:  1024 * 1024,
		AllowedOrigins:    []string{"*"},
	}

	server := httpserver.New(appCfg, srvCfg)
	assert.NotNil(t, server)
	assert.NotNil(t, server.Engine)

	cleanupCalled := false
	server.RegisterCleanup(func(ctx context.Context) error {
		cleanupCalled = true
		return nil
	})

	assert.False(t, cleanupCalled)
}
