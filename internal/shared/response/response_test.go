package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestResponse_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.OK(c, "Fetched items successfully", map[string]string{"id": "1"})

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Fetched items successfully", resp.Message)
	assert.NotNil(t, resp.Data)
}

func TestResponse_Error(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := appErrors.NewValidationError("invalid payload", appErrors.FieldError{Field: "username", Message: "required"})
	response.Error(c, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp response.APIResponse
	unmarshalErr := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, unmarshalErr)
	assert.False(t, resp.Success)
	assert.Equal(t, "invalid payload", resp.Message)
	assert.NotNil(t, resp.Errors)
}
