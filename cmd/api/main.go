package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	articleApp "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application"
	articleHTTP "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/delivery/http"
	articlePersistence "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/infrastructure/persistence"
	authApp "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application"
	authHTTP "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/delivery/http"
	authHandler "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/delivery/http/handler"
	authJWT "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/jwt"
	authPassword "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/password"
	authPersistence "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/persistence"
	healthApp "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/application"
	healthHTTP "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/delivery/http"
	healthInfra "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/infrastructure"
	rbacApp "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/application"
	rbacHTTP "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/delivery/http"
	rbacCache "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/infrastructure/cache"
	rbacPersistence "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/infrastructure/persistence"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/database"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/httpserver"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/metrics"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/redis"
	goredis "github.com/redis/go-redis/v9"
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

	// 3. Initialize MySQL Database Pool (Fail-fast on initialization error)
	mysqlDB, err := database.NewMySQL(cfg.MySQL)
	if err != nil {
		appLogger.Error("FATAL: Failed to initialize MySQL connection pool", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if pingErr := mysqlDB.Ping(context.Background()); pingErr != nil {
		appLogger.Warn("MySQL ping check failed during startup (will be verified by readiness probes)", slog.String("error", pingErr.Error()))
	} else {
		appLogger.Info("Connected to MySQL database successfully")
	}

	// 4. Initialize Redis Client Pool
	redisClient, err := redis.NewRedis(cfg.Redis)
	if err != nil {
		appLogger.Error("Failed to initialize Redis client", slog.String("error", err.Error()))
	} else {
		if pingErr := redisClient.Ping(context.Background()); pingErr != nil {
			appLogger.Warn("Redis ping check failed during startup (will be verified by readiness probes)", slog.String("error", pingErr.Error()))
		} else {
			appLogger.Info("Connected to Redis server successfully")
		}
	}

	// 5. Initialize Infrastructure Layer (Auth & RBAC)
	pwdHasher := authPassword.NewBcryptHasher()
	jwtService := authJWT.NewJWTService(cfg.JWT)

	var userRepo = authPersistence.NewMySQLUserRepository(mysqlDB.DB)
	var tokenRepo = authPersistence.NewRedisTokenRepository(redisClient.Client)

	// RBAC Infrastructure
	roleRepo := rbacPersistence.NewMySQLRoleRepository(mysqlDB.DB)
	permissionRepo := rbacPersistence.NewMySQLPermissionRepository(mysqlDB.DB)
	permCache := rbacCache.NewRedisPermissionCache(redisClient.Client)

	// Article Infrastructure
	articleRepo := articlePersistence.NewMySQLArticleRepository(mysqlDB.DB)

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
	rbacService := rbacApp.NewRBACService(roleRepo, permissionRepo, permCache)
	articleService := articleApp.NewArticleService(articleRepo)

	// Run Idempotent RBAC Database Seeder on Startup
	seeder := rbacPersistence.NewSeeder(roleRepo, permissionRepo)
	if seedErr := seeder.Seed(context.Background()); seedErr != nil {
		appLogger.Warn("Database seeder failed during startup (will retry on next boot)", slog.String("error", seedErr.Error()))
	}

	// 7. Initialize Delivery Handlers
	healthHdlr := healthHTTP.NewHandler(healthService)
	authHdlr := authHandler.NewAuthHandler(authService)

	// 8. Initialize HTTP Server & Routes
	server := httpserver.New(cfg.App, cfg.Server)

	// Register teardown cleanups (executed in order after HTTP server drains)
	server.RegisterCleanup(func(ctx context.Context) error {
		appLogger.Info("Closing MySQL connection pool...")
		return mysqlDB.Close()
	})
	if redisClient != nil {
		server.RegisterCleanup(func(ctx context.Context) error {
			appLogger.Info("Closing Redis client...")
			return redisClient.Close()
		})
	}

	// Register Health & Operational Routes
	healthHdlr.RegisterRoutes(server.Engine)
	metrics.RegisterDBStats(mysqlDB.DB)
	server.Engine.GET("/metrics", metrics.Handler())

	// Register API v1 Routes
	var redisClientRaw *goredis.Client
	if redisClient != nil {
		redisClientRaw = redisClient.Client
	}

	apiV1 := server.Engine.Group("/api/v1")
	if cfg.RateLimit.Enabled {
		apiV1.Use(middleware.RateLimiter(redisClientRaw, cfg.RateLimit.GeneralLimit, "general"))
	}
	apiV1.Use(middleware.Idempotency(redisClientRaw, cfg.RateLimit.IdempotencyTTL))

	// Initialize user activity tracking middleware
	activityMiddleware := middleware.UserActivityTracker(
		userRepo,
		redisClientRaw,
		cfg.UserActivity.UpdateInterval,
		cfg.UserActivity.Enabled,
	)

	{
		authGroup := apiV1.Group("")
		if cfg.RateLimit.Enabled {
			authGroup.Use(middleware.RateLimiter(redisClientRaw, cfg.RateLimit.AuthLimit, "auth"))
		}
		authHTTP.RegisterRoutes(authGroup, authHdlr, jwtService, activityMiddleware)

		rbacHTTP.RegisterRoutes(apiV1, jwtService, rbacService, activityMiddleware)
		articleHTTP.RegisterRoutes(apiV1, jwtService, rbacService, articleService, activityMiddleware)
	}

	// 9. Run HTTP Server with Graceful Shutdown
	if err := server.Run(); err != nil {
		appLogger.Error("HTTP Server terminated with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	appLogger.Info("Application shutdown completed clean")
}
