package repository

import (
	"context"
	"time"
)

// TokenRepository defines operations for managing refresh tokens / session state.
type TokenRepository interface {
	StoreRefreshToken(ctx context.Context, tokenID string, userID string, ttl time.Duration) error
	ValidateRefreshToken(ctx context.Context, tokenID string) (string, error)
	RevokeRefreshToken(ctx context.Context, tokenID string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
}
