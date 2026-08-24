package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

type CleanupFunc func(ctx context.Context) error

type Server struct {
	Engine   *gin.Engine
	http     *http.Server
	cfg      config.ServerConfig
	cleanups []CleanupFunc
}

// New creates and configures a new HTTP Server with hardened middlewares and timeouts.
func New(appCfg config.AppConfig, srvCfg config.ServerConfig) *Server {
	if appCfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	// Register hardened foundational middlewares
	engine.Use(
		middleware.SecurityHeaders(),
		middleware.MaxBodySize(srvCfg.MaxBodySizeBytes),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS(srvCfg.AllowedOrigins),
		middleware.Timeout(srvCfg.RequestTimeout),
	)

	readHeaderTimeout := srvCfg.ReadHeaderTimeout
	if readHeaderTimeout <= 0 {
		readHeaderTimeout = srvCfg.ReadTimeout
	}

	httpServer := &http.Server{
		Addr:              srvCfg.Address(),
		Handler:           engine,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       srvCfg.ReadTimeout,
		WriteTimeout:      srvCfg.WriteTimeout,
		IdleTimeout:       srvCfg.IdleTimeout,
	}

	return &Server{
		Engine:   engine,
		http:     httpServer,
		cfg:      srvCfg,
		cleanups: make([]CleanupFunc, 0),
	}
}

// RegisterCleanup registers a teardown function to run after HTTP requests drain.
func (s *Server) RegisterCleanup(fn CleanupFunc) {
	s.cleanups = append(s.cleanups, fn)
}

// Run starts the HTTP server in a goroutine and coordinates graceful shutdown.
func (s *Server) Run() error {
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	serverErr := make(chan error, 1)
	go func() {
		logger.Get().Info(fmt.Sprintf("HTTP server starting on %s", s.cfg.Address()))
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdownChan:
		logger.Get().Info(fmt.Sprintf("Received shutdown signal: %v. Initiating graceful shutdown...", sig))
		ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()

		// 1. Drain active HTTP requests
		if err := s.http.Shutdown(ctx); err != nil {
			logger.Get().Error("HTTP server forced to shutdown with error", "error", err)
		} else {
			logger.Get().Info("HTTP server gracefully drained and stopped")
		}

		// 2. Execute resource cleanups (MySQL, Redis, etc.)
		for _, cleanup := range s.cleanups {
			if err := cleanup(ctx); err != nil {
				logger.Get().Error("Cleanup task failed during shutdown", "error", err)
			}
		}
	}

	return nil
}
