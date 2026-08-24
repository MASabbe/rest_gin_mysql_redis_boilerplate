package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	healthHTTP "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/delivery/http"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/health/domain"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockHealthService struct {
	liveReport  domain.HealthReport
	readyReport domain.HealthReport
	overReport  domain.HealthReport
}

func (m *mockHealthService) CheckLiveness(ctx context.Context) domain.HealthReport {
	return m.liveReport
}
func (m *mockHealthService) CheckReadiness(ctx context.Context) domain.HealthReport {
	return m.readyReport
}
func (m *mockHealthService) CheckOverall(ctx context.Context) domain.HealthReport {
	return m.overReport
}

func setupTestRouter(svc *mockHealthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := healthHTTP.NewHandler(svc)
	handler.RegisterRoutes(r)
	return r
}

func TestHealthHandler_Live(t *testing.T) {
	svc := &mockHealthService{
		liveReport: domain.HealthReport{Status: domain.StatusUP},
	}
	r := setupTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health/live", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestHealthHandler_Ready_Success(t *testing.T) {
	svc := &mockHealthService{
		readyReport: domain.HealthReport{Status: domain.StatusUP},
	}
	r := setupTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health/ready", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthHandler_Ready_Unavailable(t *testing.T) {
	svc := &mockHealthService{
		readyReport: domain.HealthReport{Status: domain.StatusDOWN},
	}
	r := setupTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health/ready", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
