package infrastructure

import (
	"context"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/domain"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/database"
)

type MySQLChecker struct {
	mysql *database.MySQL
}

// NewMySQLChecker creates a Checker for MySQL.
func NewMySQLChecker(mysql *database.MySQL) domain.Checker {
	return &MySQLChecker{mysql: mysql}
}

func (c *MySQLChecker) Name() string {
	return "mysql"
}

func (c *MySQLChecker) Check(ctx context.Context) domain.ComponentCheck {
	start := time.Now()
	if c.mysql == nil {
		return domain.ComponentCheck{
			Name:      c.Name(),
			Status:    domain.StatusDOWN,
			Error:     "mysql instance is not configured",
			Latency:   time.Since(start),
			LatencyMs: float64(time.Since(start).Microseconds()) / 1000.0,
		}
	}

	err := c.mysql.Ping(ctx)
	latency := time.Since(start)
	latencyMs := float64(latency.Microseconds()) / 1000.0

	if err != nil {
		return domain.ComponentCheck{
			Name:      c.Name(),
			Status:    domain.StatusDOWN,
			Error:     err.Error(),
			Latency:   latency,
			LatencyMs: latencyMs,
		}
	}

	return domain.ComponentCheck{
		Name:      c.Name(),
		Status:    domain.StatusUP,
		Details:   "database connection responsive",
		Latency:   latency,
		LatencyMs: latencyMs,
	}
}
