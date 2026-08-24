package application

import (
	"context"
	"sync"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/domain"
)

type HealthService interface {
	CheckLiveness(ctx context.Context) domain.HealthReport
	CheckReadiness(ctx context.Context) domain.HealthReport
	CheckOverall(ctx context.Context) domain.HealthReport
}

type healthService struct {
	checkers []domain.Checker
}

// NewHealthService creates a new HealthService instance.
func NewHealthService(checkers ...domain.Checker) HealthService {
	return &healthService{
		checkers: checkers,
	}
}

// CheckLiveness performs a shallow check ensuring the process is responding.
func (s *healthService) CheckLiveness(ctx context.Context) domain.HealthReport {
	return domain.HealthReport{
		Status:    domain.StatusUP,
		Timestamp: time.Now().UTC(),
	}
}

// CheckReadiness executes registered infrastructure checkers concurrently and returns system readiness.
func (s *healthService) CheckReadiness(ctx context.Context) domain.HealthReport {
	return s.runChecks(ctx)
}

// CheckOverall returns a comprehensive system health report.
func (s *healthService) CheckOverall(ctx context.Context) domain.HealthReport {
	return s.runChecks(ctx)
}

func (s *healthService) runChecks(ctx context.Context) domain.HealthReport {
	components := make(map[string]domain.ComponentCheck)
	if len(s.checkers) == 0 {
		return domain.HealthReport{
			Status:     domain.StatusUP,
			Timestamp:  time.Now().UTC(),
			Components: components,
		}
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	overallStatus := domain.StatusUP

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	for _, checker := range s.checkers {
		wg.Add(1)
		go func(c domain.Checker) {
			defer wg.Done()
			res := c.Check(ctxWithTimeout)

			mu.Lock()
			components[c.Name()] = res
			if res.Status == domain.StatusDOWN {
				overallStatus = domain.StatusDOWN
			} else if res.Status == domain.StatusDEGRADED && overallStatus != domain.StatusDOWN {
				overallStatus = domain.StatusDEGRADED
			}
			mu.Unlock()
		}(checker)
	}

	wg.Wait()

	return domain.HealthReport{
		Status:     overallStatus,
		Timestamp:  time.Now().UTC(),
		Components: components,
	}
}
