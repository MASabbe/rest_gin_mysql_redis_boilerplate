package command_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application/command"
	"github.com/stretchr/testify/assert"
)

func TestCommands_Validation(t *testing.T) {
	// RegisterCommand
	regValid := command.RegisterCommand{Email: "user@test.com", Password: "Password123!"}
	assert.NoError(t, regValid.Validate())

	regInvalid := command.RegisterCommand{Email: "", Password: "short"}
	assert.Error(t, regInvalid.Validate())

	// LoginCommand
	loginValid := command.LoginCommand{Email: "user@test.com", Password: "Password123!"}
	assert.NoError(t, loginValid.Validate())

	loginInvalid := command.LoginCommand{Email: "", Password: ""}
	assert.Error(t, loginInvalid.Validate())

	// RefreshTokenCommand
	refreshValid := command.RefreshTokenCommand{RefreshToken: "valid-token"}
	assert.NoError(t, refreshValid.Validate())

	refreshInvalid := command.RefreshTokenCommand{RefreshToken: ""}
	assert.Error(t, refreshInvalid.Validate())

	// LogoutCommand
	logoutValid := command.LogoutCommand{RefreshToken: "valid-token"}
	assert.NoError(t, logoutValid.Validate())

	logoutInvalid := command.LogoutCommand{RefreshToken: ""}
	assert.Error(t, logoutInvalid.Validate())
}
