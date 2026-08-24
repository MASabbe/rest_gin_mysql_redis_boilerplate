package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/config"
	_ "github.com/go-sql-driver/mysql"
)

// MySQL holds the database connection pool.
type MySQL struct {
	DB *sql.DB
}

// NewMySQL creates a new MySQL connection pool from the provided configuration.
func NewMySQL(cfg config.MySQLConfig) (*MySQL, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return &MySQL{DB: db}, nil
}

// Ping verifies database connectivity with context timeout.
func (m *MySQL) Ping(ctx context.Context) error {
	if m == nil || m.DB == nil {
		return fmt.Errorf("mysql connection is uninitialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return m.DB.PingContext(ctx)
}

// Close gracefully closes the MySQL database connection pool.
func (m *MySQL) Close() error {
	if m == nil || m.DB == nil {
		return nil
	}
	return m.DB.Close()
}
