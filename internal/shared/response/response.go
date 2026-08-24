package response

import (
	"net/http"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
	"github.com/gin-gonic/gin"
)

// APIResponse represents the standardized JSON response contract.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Errors  any    `json:"errors,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

// Success writes a successful response with custom status code, data, and optional metadata.
func Success(c *gin.Context, httpStatus int, message string, data any, meta ...any) {
	if message == "" {
		message = "Success"
	}
	var metaData any
	if len(meta) > 0 {
		metaData = meta[0]
	}

	c.JSON(httpStatus, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    metaData,
	})
}

// OK writes a 200 OK successful response.
func OK(c *gin.Context, message string, data any) {
	Success(c, http.StatusOK, message, data)
}

// Created writes a 201 Created response.
func Created(c *gin.Context, message string, data any) {
	Success(c, http.StatusCreated, message, data)
}

// Error writes a standardized error response mapped from an error or *appErrors.AppError.
func Error(c *gin.Context, err error) {
	appErr := appErrors.AsAppError(err)

	// Log internal or infrastructure errors for observability
	if appErr.HTTPStatus >= http.StatusInternalServerError {
		logger.WithContext(c.Request.Context()).Error("Server error encountered",
			"code", appErr.Code,
			"message", appErr.Message,
			"error", appErr.Err,
		)
	}

	var errorsList any
	if len(appErr.Details) > 0 {
		errorsList = appErr.Details
	}

	c.JSON(appErr.HTTPStatus, APIResponse{
		Success: false,
		Message: appErr.Message,
		Data:    nil,
		Errors:  errorsList,
		Meta:    nil,
	})
}
