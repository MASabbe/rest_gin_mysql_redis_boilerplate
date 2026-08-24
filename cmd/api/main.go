package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	authApp "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application"
	authHTTP "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/delivery/http"
	authHandler "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/delivery/http/handler"
	authJWT "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/jwt"
	authPassword "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/password"
	authPersistence "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/persistence"
	healthApp "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/application"
	healthHTTP "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/delivery/http"
	healthInfra "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/infrastructure"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/database"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/httpserver"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/redis"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize structured logger
	appLogger := logger.Init(cfg.App.Env, cfg.App.LogLevel)
	appLogger.Info("Initializing backend service...",
		slog.String("app_name", cfg.App.Name),
		slog.String("env", cfg.App.Env),
		slog.String("port", cfg.Server.Port),
	)

	// 3. Initialize MySQL Database Pool
	mysqlDB, err := database.NewMySQL(cfg.MySQL)
	if err != nil {
		appLogger.Error("Failed to initialize MySQL connection pool", slog.String("error", err.Error()))
	} else {
		defer func() {
			if closeErr := mysqlDB.Close(); closeErr != nil {
				appLogger.Error("Error closing MySQL connection", slog.String("error", closeErr.Error()))
			}
		}()

		if pingErr := mysqlDB.Ping(context.Background()); pingErr != nil {
			appLogger.Warn("MySQL ping check failed during startup (will be verified by readiness probes)", slog.String("error", pingErr.Error()))
		} else {
			appLogger.Info("Connected to MySQL database successfully")
		}
	}

	// 4. Initialize Redis Client Pool
	redisClient, err := redis.NewRedis(cfg.Redis)
	if err != nil {
		appLogger.Error("Failed to initialize Redis client", slog.String("error", err.Error()))
	} else {
		defer func() {
			if closeErr := redisClient.Close(); closeErr != nil {
				appLogger.Error("Error closing Redis client", slog.String("error", closeErr.Error()))
			}
		}()

		if pingErr := redisClient.Ping(context.Background()); pingErr != nil {
			appLogger.Warn("Redis ping check failed during startup (will be verified by readiness probes)", slog.String("error", pingErr.Error()))
		} else {
			appLogger.Info("Connected to Redis server successfully")
		}
	}

	// 5. Initialize Infrastructure Layer
	pwdHasher := authPassword.NewBcryptHasher()
	jwtService := authJWT.NewJWTService(cfg.JWT)

	var userRepo = authPersistence.NewMySQLUserRepository(mysqlDB.DB)
	var tokenRepo = authPersistence.NewRedisTokenRepository(redisClient.Client)

	// Health Checkers
	mysqlChecker := healthInfra.NewMySQLChecker(mysqlDB)
	redisChecker := healthInfra.NewRedisChecker(redisClient)

	// 6. Initialize Application Services
	healthService := healthApp.NewHealthService(mysqlChecker, redisChecker)
	authService := authApp.NewAuthService(
		userRepo,
		tokenRepo,
		pwdHasher,
		jwtService,
		cfg.JWT.RefreshTokenTTL,
	)

	// 7. Initialize Delivery Handlers
	healthHdlr := healthHTTP.NewHandler(healthService)
	authHdlr := authHandler.NewAuthHandler(authService)

	// 8. Initialize HTTP Server & Routes
	server := httpserver.New(cfg.App, cfg.Server)

	// Register Health Routes
	healthHdlr.RegisterRoutes(server.Engine)

	// Register API v1 Routes
	apiV1 := server.Engine.Group("/api/v1")
	{
		authHTTP.RegisterRoutes(apiV1, authHdlr, jwtService)
	}

	// 9. Run HTTP Server with Graceful Shutdown
	if err := server.Run(); err != nil {
		appLogger.Error("HTTP Server terminated with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	appLogger.Info("Application shutdown completed clean")
}
