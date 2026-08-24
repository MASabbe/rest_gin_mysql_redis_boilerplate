package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/domain"
	"github.com/stretchr/testify/assert"
)

type mockChecker struct {
	name   string
	status domain.Status
	err    string
}

func (m *mockChecker) Name() string {
	return m.name
}

func (m *mockChecker) Check(ctx context.Context) domain.ComponentCheck {
	return domain.ComponentCheck{
		Name:    m.name,
		Status:  m.status,
		Error:   m.err,
		Latency: 10 * time.Millisecond,
	}
}

func TestHealthService_Liveness(t *testing.T) {
	svc := application.NewHealthService()
	report := svc.CheckLiveness(context.Background())

	assert.Equal(t, domain.StatusUP, report.Status)
	assert.Empty(t, report.Components)
}

func TestHealthService_Readiness_AllHealthy(t *testing.T) {
	c1 := &mockChecker{name: "mysql", status: domain.StatusUP}
	c2 := &mockChecker{name: "redis", status: domain.StatusUP}

	svc := application.NewHealthService(c1, c2)
	report := svc.CheckReadiness(context.Background())

	assert.Equal(t, domain.StatusUP, report.Status)
	assert.Len(t, report.Components, 2)
	assert.Equal(t, domain.StatusUP, report.Components["mysql"].Status)
	assert.Equal(t, domain.StatusUP, report.Components["redis"].Status)
}

func TestHealthService_Readiness_OneDown(t *testing.T) {
	c1 := &mockChecker{name: "mysql", status: domain.StatusUP}
	c2 := &mockChecker{name: "redis", status: domain.StatusDOWN, err: "connection refused"}

	svc := application.NewHealthService(c1, c2)
	report := svc.CheckReadiness(context.Background())

	assert.Equal(t, domain.StatusDOWN, report.Status)
	assert.Equal(t, domain.StatusDOWN, report.Components["redis"].Status)
}
