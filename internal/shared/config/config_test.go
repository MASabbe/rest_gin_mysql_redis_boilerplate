package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Load_DefaultAndValidation(t *testing.T) {
	// Clear any overrides
	os.Clearenv()

	cfg, err := config.Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "backend-api", cfg.App.Name)
	assert.Equal(t, "development", cfg.App.Env)
	assert.False(t, cfg.App.IsProduction())
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "0.0.0.0:8080", cfg.Server.Address())
	assert.Equal(t, "127.0.0.1:6379", cfg.Redis.Address())
	assert.Contains(t, cfg.MySQL.DSN(), "root:@tcp(127.0.0.1:3306)/app_db")
	assert.Equal(t, 5*time.Second, cfg.Server.ReadHeaderTimeout)
	assert.Equal(t, int64(2*1024*1024), cfg.Server.MaxBodySizeBytes)
	assert.True(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 100, cfg.RateLimit.GeneralLimit)
	assert.Equal(t, 10, cfg.RateLimit.AuthLimit)
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTokenTTL)
	assert.Equal(t, 7*24*time.Hour, cfg.JWT.RefreshTokenTTL)
}

func TestConfig_Validation_FailShortSecret(t *testing.T) {
	os.Clearenv()
	t.Setenv("JWT_SECRET", "short")

	_, err := config.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET must be at least 32 characters")
}

func TestConfig_IsProduction(t *testing.T) {
	cfg := config.Config{
		App: config.AppConfig{
			Env: "production",
		},
	}
	assert.True(t, cfg.App.IsProduction())
}
