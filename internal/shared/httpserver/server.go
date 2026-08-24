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

type Server struct {
	Engine *gin.Engine
	http   *http.Server
	cfg    config.ServerConfig
}

// New creates and configures a new HTTP Server with default middlewares.
func New(appCfg config.AppConfig, srvCfg config.ServerConfig) *Server {
	if appCfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	// Register foundational middlewares
	engine.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS(srvCfg.AllowedOrigins),
		middleware.Timeout(srvCfg.RequestTimeout),
	)

	httpServer := &http.Server{
		Addr:         srvCfg.Address(),
		Handler:      engine,
		ReadTimeout:  srvCfg.ReadTimeout,
		WriteTimeout: srvCfg.WriteTimeout,
		IdleTimeout:  srvCfg.IdleTimeout,
	}

	return &Server{
		Engine: engine,
		http:   httpServer,
		cfg:    srvCfg,
	}
}

// Run starts the HTTP server in a goroutine and waits for shutdown signals.
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

		if err := s.http.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shutdown with error: %w", err)
		}
		logger.Get().Info("HTTP server gracefully stopped")
	}

	return nil
}
