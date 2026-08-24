package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the category of the application error.
type ErrorType string

const (
	TypeValidation     ErrorType = "VALIDATION_ERROR"
	TypeUnauthorized   ErrorType = "UNAUTHORIZED"
	TypeForbidden      ErrorType = "FORBIDDEN"
	TypeNotFound       ErrorType = "NOT_FOUND"
	TypeConflict       ErrorType = "CONFLICT"
	TypeBusiness       ErrorType = "BUSINESS_ERROR"
	TypeInternal       ErrorType = "INTERNAL_ERROR"
	TypeInfrastructure ErrorType = "INFRASTRUCTURE_ERROR"
)

// FieldError represents a specific field-level validation issue.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// AppError is the central structured application error.
type AppError struct {
	Type       ErrorType    `json:"type"`
	Code       string       `json:"code"`
	Message    string       `json:"message"`
	HTTPStatus int          `json:"-"`
	Details    []FieldError `json:"details,omitempty"`
	Err        error        `json:"-"` // Underlying cause for internal logging only
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a 400 Bad Request error.
func NewValidationError(message string, details ...FieldError) *AppError {
	if message == "" {
		message = "Validation failed"
	}
	return &AppError{
		Type:       TypeValidation,
		Code:       "VALIDATION_ERROR",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Details:    details,
	}
}

// NewUnauthorizedError creates a 401 Unauthorized error.
func NewUnauthorizedError(message string, err ...error) *AppError {
	if message == "" {
		message = "Authentication required or invalid"
	}
	var underlying error
	if len(err) > 0 {
		underlying = err[0]
	}
	return &AppError{
		Type:       TypeUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
		Err:        underlying,
	}
}

// NewForbiddenError creates a 403 Forbidden error.
func NewForbiddenError(message string, err ...error) *AppError {
	if message == "" {
		message = "Access forbidden"
	}
	var underlying error
	if len(err) > 0 {
		underlying = err[0]
	}
	return &AppError{
		Type:       TypeForbidden,
		Code:       "FORBIDDEN",
		Message:    message,
		HTTPStatus: http.StatusForbidden,
		Err:        underlying,
	}
}

// NewNotFoundError creates a 404 Not Found error.
func NewNotFoundError(message string, err ...error) *AppError {
	if message == "" {
		message = "Requested resource not found"
	}
	var underlying error
	if len(err) > 0 {
		underlying = err[0]
	}
	return &AppError{
		Type:       TypeNotFound,
		Code:       "NOT_FOUND",
		Message:    message,
		HTTPStatus: http.StatusNotFound,
		Err:        underlying,
	}
}

// NewConflictError creates a 409 Conflict error.
func NewConflictError(message string, err ...error) *AppError {
	if message == "" {
		message = "Resource conflict detected"
	}
	var underlying error
	if len(err) > 0 {
		underlying = err[0]
	}
	return &AppError{
		Type:       TypeConflict,
		Code:       "CONFLICT",
		Message:    message,
		HTTPStatus: http.StatusConflict,
		Err:        underlying,
	}
}

// NewBusinessError creates a 422 Unprocessable Entity business error.
func NewBusinessError(code, message string, err ...error) *AppError {
	if code == "" {
		code = "BUSINESS_RULE_VIOLATION"
	}
	var underlying error
	if len(err) > 0 {
		underlying = err[0]
	}
	return &AppError{
		Type:       TypeBusiness,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusUnprocessableEntity,
		Err:        underlying,
	}
}

// NewInternalError creates a 500 Internal Server Error.
func NewInternalError(message string, err ...error) *AppError {
	if message == "" {
		message = "An unexpected internal error occurred"
	}
	var underlying error
	if len(err) > 0 {
		underlying = err[0]
	}
	return &AppError{
		Type:       TypeInternal,
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        underlying,
	}
}

// NewInfrastructureError creates a 503 Service Unavailable error.
func NewInfrastructureError(message string, err ...error) *AppError {
	if message == "" {
		message = "External service or database temporarily unavailable"
	}
	var underlying error
	if len(err) > 0 {
		underlying = err[0]
	}
	return &AppError{
		Type:       TypeInfrastructure,
		Code:       "SERVICE_UNAVAILABLE",
		Message:    message,
		HTTPStatus: http.StatusServiceUnavailable,
		Err:        underlying,
	}
}

// AsAppError attempts to cast or convert any error into an *AppError.
func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return NewInternalError("An unexpected error occurred", err)
}
