package errors_test

import (
	"errors"
	"net/http"
	"testing"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/stretchr/testify/assert"
)

func TestAppError_Constructors(t *testing.T) {
	valErr := appErrors.NewValidationError("invalid input", appErrors.FieldError{Field: "email", Message: "required"})
	assert.Equal(t, http.StatusBadRequest, valErr.HTTPStatus)
	assert.Equal(t, appErrors.TypeValidation, valErr.Type)
	assert.Len(t, valErr.Details, 1)
	assert.Contains(t, valErr.Error(), "invalid input")

	unauthErr := appErrors.NewUnauthorizedError("invalid token")
	assert.Equal(t, http.StatusUnauthorized, unauthErr.HTTPStatus)

	forbidErr := appErrors.NewForbiddenError("forbidden access")
	assert.Equal(t, http.StatusForbidden, forbidErr.HTTPStatus)

	notFoundErr := appErrors.NewNotFoundError("user not found")
	assert.Equal(t, http.StatusNotFound, notFoundErr.HTTPStatus)

	conflictErr := appErrors.NewConflictError("email already registered")
	assert.Equal(t, http.StatusConflict, conflictErr.HTTPStatus)

	payloadErr := appErrors.NewPayloadTooLargeError("file too big")
	assert.Equal(t, http.StatusRequestEntityTooLarge, payloadErr.HTTPStatus)
	assert.Equal(t, appErrors.TypePayloadTooLarge, payloadErr.Type)

	rateErr := appErrors.NewRateLimitExceededError("too fast")
	assert.Equal(t, http.StatusTooManyRequests, rateErr.HTTPStatus)
	assert.Equal(t, appErrors.TypeRateLimitExceeded, rateErr.Type)

	bizErr := appErrors.NewBusinessError("INSUFFICIENT_BALANCE", "balance too low")
	assert.Equal(t, http.StatusUnprocessableEntity, bizErr.HTTPStatus)

	infraErr := appErrors.NewInfrastructureError("db unreachable")
	assert.Equal(t, http.StatusServiceUnavailable, infraErr.HTTPStatus)

	internalErr := appErrors.NewInternalError("panic occurred", errors.New("raw db error"))
	assert.Equal(t, http.StatusInternalServerError, internalErr.HTTPStatus)
	assert.Equal(t, "raw db error", internalErr.Unwrap().Error())
	assert.Contains(t, internalErr.Error(), "raw db error")
}

func TestAsAppError(t *testing.T) {
	assert.Nil(t, appErrors.AsAppError(nil))

	customErr := appErrors.NewNotFoundError("not found")
	assert.Equal(t, customErr, appErrors.AsAppError(customErr))

	rawErr := errors.New("something went wrong")
	converted := appErrors.AsAppError(rawErr)
	assert.Equal(t, http.StatusInternalServerError, converted.HTTPStatus)
	assert.Equal(t, rawErr, converted.Unwrap())
}
