package application

import (
	"context"
	"strings"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application/dto"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/repository"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/service"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/metrics"
)

type AuthService interface {
	Register(ctx context.Context, cmd command.RegisterCommand) (*dto.AuthResultDTO, error)
	Login(ctx context.Context, cmd command.LoginCommand) (*dto.AuthResultDTO, error)
	RefreshToken(ctx context.Context, cmd command.RefreshTokenCommand) (*dto.TokenPairDTO, error)
	Logout(ctx context.Context, cmd command.LogoutCommand) error
	GetMe(ctx context.Context, query query.GetUserByIDQuery) (*dto.UserDTO, error)
}

type authService struct {
	userRepo        repository.UserRepository
	tokenRepo       repository.TokenRepository
	passwordHasher  service.PasswordHasher
	tokenService    service.TokenService
	refreshTokenTTL time.Duration
}

// NewAuthService constructs a new AuthService application service.
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	passwordHasher service.PasswordHasher,
	tokenService service.TokenService,
	refreshTokenTTL time.Duration,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		passwordHasher:  passwordHasher,
		tokenService:    tokenService,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *authService) Register(ctx context.Context, cmd command.RegisterCommand) (*dto.AuthResultDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	cleanEmail := strings.TrimSpace(strings.ToLower(cmd.Email))

	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, cleanEmail)
	if err != nil && !isNotFoundError(err) {
		return nil, err
	}
	if existingUser != nil {
		return nil, appErrors.NewConflictError("email already registered")
	}

	// Hash password
	hashedPassword, err := s.passwordHasher.Hash(cmd.Password)
	if err != nil {
		return nil, err
	}

	// Create user entity
	user, err := entity.NewUser(cleanEmail, hashedPassword)
	if err != nil {
		return nil, err
	}

	// Persist user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Issue token pair
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Store refresh token in store
	if s.tokenRepo != nil {
		if err := s.tokenRepo.StoreRefreshToken(ctx, tokenPair.RefreshTokenID, user.ID, s.refreshTokenTTL); err != nil {
			return nil, err
		}
	}

	return &dto.AuthResultDTO{
		User: dto.ToUserDTO(user),
		Tokens: dto.TokenPairDTO{
			AccessToken:           tokenPair.AccessToken,
			RefreshToken:          tokenPair.RefreshToken,
			TokenType:             tokenPair.TokenType,
			AccessTokenExpiresIn:  tokenPair.AccessTokenExpiresIn,
			RefreshTokenExpiresIn: tokenPair.RefreshTokenExpiresIn,
		},
	}, nil
}

func (s *authService) Login(ctx context.Context, cmd command.LoginCommand) (*dto.AuthResultDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	cleanEmail := strings.TrimSpace(strings.ToLower(cmd.Email))

	user, err := s.userRepo.FindByEmail(ctx, cleanEmail)
	if err != nil {
		if isNotFoundError(err) {
			metrics.RecordAuthFailure("user_not_found")
			return nil, appErrors.NewUnauthorizedError("invalid email or password")
		}
		return nil, err
	}

	if err := s.passwordHasher.Compare(user.PasswordHash, cmd.Password); err != nil {
		metrics.RecordAuthFailure("invalid_password")
		return nil, appErrors.NewUnauthorizedError("invalid email or password")
	}

	if !user.IsActive() {
		metrics.RecordAuthFailure("account_inactive")
		return nil, appErrors.NewForbiddenError("account is inactive or suspended")
	}

	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	if s.tokenRepo != nil {
		if err := s.tokenRepo.StoreRefreshToken(ctx, tokenPair.RefreshTokenID, user.ID, s.refreshTokenTTL); err != nil {
			return nil, err
		}
	}

	return &dto.AuthResultDTO{
		User: dto.ToUserDTO(user),
		Tokens: dto.TokenPairDTO{
			AccessToken:           tokenPair.AccessToken,
			RefreshToken:          tokenPair.RefreshToken,
			TokenType:             tokenPair.TokenType,
			AccessTokenExpiresIn:  tokenPair.AccessTokenExpiresIn,
			RefreshTokenExpiresIn: tokenPair.RefreshTokenExpiresIn,
		},
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, cmd command.RefreshTokenCommand) (*dto.TokenPairDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	claims, err := s.tokenService.ValidateRefreshToken(cmd.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Verify token in store
	if s.tokenRepo != nil {
		storedUserID, err := s.tokenRepo.ValidateRefreshToken(ctx, claims.TokenID)
		if err != nil {
			return nil, err
		}
		if storedUserID != claims.UserID {
			return nil, appErrors.NewUnauthorizedError("invalid token ownership")
		}

		// Token rotation: revoke old refresh token
		_ = s.tokenRepo.RevokeRefreshToken(ctx, claims.TokenID)
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, appErrors.NewUnauthorizedError("user not found or invalid", err)
	}

	if !user.IsActive() {
		return nil, appErrors.NewForbiddenError("account is inactive or suspended")
	}

	// Generate new token pair
	newPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	if s.tokenRepo != nil {
		if err := s.tokenRepo.StoreRefreshToken(ctx, newPair.RefreshTokenID, user.ID, s.refreshTokenTTL); err != nil {
			return nil, err
		}
	}

	return &dto.TokenPairDTO{
		AccessToken:           newPair.AccessToken,
		RefreshToken:          newPair.RefreshToken,
		TokenType:             newPair.TokenType,
		AccessTokenExpiresIn:  newPair.AccessTokenExpiresIn,
		RefreshTokenExpiresIn: newPair.RefreshTokenExpiresIn,
	}, nil
}

func (s *authService) Logout(ctx context.Context, cmd command.LogoutCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	claims, err := s.tokenService.ValidateRefreshToken(cmd.RefreshToken)
	if err != nil {
		return err
	}

	if s.tokenRepo != nil {
		return s.tokenRepo.RevokeRefreshToken(ctx, claims.TokenID)
	}

	return nil
}

func (s *authService) GetMe(ctx context.Context, q query.GetUserByIDQuery) (*dto.UserDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	userDTO := dto.ToUserDTO(user)
	return &userDTO, nil
}

func isNotFoundError(err error) bool {
	appErr := appErrors.AsAppError(err)
	return appErr != nil && appErr.Type == appErrors.TypeNotFound
}
