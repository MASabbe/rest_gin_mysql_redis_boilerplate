package handler

import (
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/delivery/http/request"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/delivery/http/response"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	sharedResponse "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService application.AuthService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService application.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "body",
			Message: err.Error(),
		}))
		return
	}

	result, err := h.authService.Register(c.Request.Context(), command.RegisterCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.Created(c, "User registered successfully", response.FromAuthResultDTO(result))
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "body",
			Message: err.Error(),
		}))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), command.LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Login successful", response.FromAuthResultDTO(result))
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "refresh_token",
			Message: "refresh_token is required",
		}))
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), command.RefreshTokenCommand{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Token refreshed successfully", response.FromTokenPairDTO(result))
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req request.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "refresh_token",
			Message: "refresh_token is required",
		}))
		return
	}

	err := h.authService.Logout(c.Request.Context(), command.LogoutCommand{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "Logged out successfully", nil)
}

// Me handles GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok || userID == "" {
		sharedResponse.Error(c, appErrors.NewUnauthorizedError("unauthorized"))
		return
	}

	user, err := h.authService.GetMe(c.Request.Context(), query.GetUserByIDQuery{
		UserID: userID,
	})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "User profile retrieved successfully", response.FromUserDTO(user))
}
