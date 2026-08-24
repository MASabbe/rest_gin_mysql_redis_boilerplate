package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the application root configuration.
type Config struct {
	App    AppConfig
	Server ServerConfig
	MySQL  MySQLConfig
	Redis  RedisConfig
	JWT    JWTConfig
}

// AppConfig represents general application settings.
type AppConfig struct {
	Name     string
	Env      string // "development", "staging", "production", "test"
	LogLevel string // "debug", "info", "warn", "error"
}

func (c AppConfig) IsProduction() bool {
	return strings.ToLower(c.Env) == "production"
}

// ServerConfig represents HTTP server settings.
type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	RequestTimeout  time.Duration
	AllowedOrigins  []string
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

// MySQLConfig represents MySQL database connection settings.
type MySQLConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	Params          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func (m MySQLConfig) DSN() string {
	params := m.Params
	if params == "" {
		params = "charset=utf8mb4&parseTime=True&loc=Local"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		m.User, m.Password, m.Host, m.Port, m.Database, params)
}

// RedisConfig represents Redis connection settings.
type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// JWTConfig represents JSON Web Token configuration.
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
	Audience        string
}

// Load reads configuration from environment variables, optionally loading from a .env file first.
func Load(envFiles ...string) (*Config, error) {
	if len(envFiles) > 0 {
		for _, f := range envFiles {
			if _, err := os.Stat(f); err == nil {
				_ = godotenv.Overload(f)
			}
		}
	} else {
		// Load standard .env if exists, ignore error if missing in prod/containers
		if _, err := os.Stat(".env"); err == nil {
			_ = godotenv.Load(".env")
		}
	}

	cfg := &Config{
		App: AppConfig{
			Name:     getEnv("APP_NAME", "backend-api"),
			Env:      getEnv("APP_ENV", "development"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			RequestTimeout:  getDurationEnv("SERVER_REQUEST_TIMEOUT", 30*time.Second),
			AllowedOrigins:  splitAndTrim(getEnv("SERVER_CORS_ALLOWED_ORIGINS", "*"), ","),
		},
		MySQL: MySQLConfig{
			Host:            getEnv("MYSQL_HOST", "127.0.0.1"),
			Port:            getEnv("MYSQL_PORT", "3306"),
			User:            getEnv("MYSQL_USER", "root"),
			Password:        getEnv("MYSQL_PASSWORD", ""),
			Database:        getEnv("MYSQL_DATABASE", "app_db"),
			Params:          getEnv("MYSQL_PARAMS", "charset=utf8mb4&parseTime=True&loc=Local"),
			MaxOpenConns:    getIntEnv("MYSQL_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("MYSQL_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDurationEnv("MYSQL_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getDurationEnv("MYSQL_CONN_MAX_IDLE_TIME", 2*time.Minute),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "127.0.0.1"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getIntEnv("REDIS_DB", 0),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 2),
			DialTimeout:  getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production-min32chars"),
			AccessTokenTTL:  getDurationEnv("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: getDurationEnv("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour),
			Issuer:          getEnv("JWT_ISSUER", "backend-api"),
			Audience:        getEnv("JWT_AUDIENCE", "backend-api-users"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate verifies that mandatory configuration fields are populated.
func (c *Config) Validate() error {
	var errs []string

	if c.App.Name == "" {
		errs = append(errs, "APP_NAME is required")
	}
	if c.App.Env == "" {
		errs = append(errs, "APP_ENV is required")
	}
	if c.Server.Port == "" {
		errs = append(errs, "SERVER_PORT is required")
	}
	if c.MySQL.Host == "" {
		errs = append(errs, "MYSQL_HOST is required")
	}
	if c.MySQL.Database == "" {
		errs = append(errs, "MYSQL_DATABASE is required")
	}
	if c.Redis.Host == "" {
		errs = append(errs, "REDIS_HOST is required")
	}
	if c.JWT.Secret == "" {
		errs = append(errs, "JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		errs = append(errs, "JWT_SECRET must be at least 32 characters long for security")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getIntEnv(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var res []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			res = append(res, trimmed)
		}
	}
	return res
}
