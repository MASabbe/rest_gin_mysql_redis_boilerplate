package query_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application/query"
	"github.com/stretchr/testify/assert"
)

func TestGetUserByIDQuery_Validation(t *testing.T) {
	valid := query.GetUserByIDQuery{UserID: "usr-123"}
	assert.NoError(t, valid.Validate())

	invalid := query.GetUserByIDQuery{UserID: ""}
	assert.Error(t, invalid.Validate())
}
