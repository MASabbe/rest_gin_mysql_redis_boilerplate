package persistence

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/repository"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
)

// InMemoryUserRepository is a thread-safe in-memory implementation of UserRepository.
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*entity.User // key: ID
}

func NewInMemoryUserRepository() repository.UserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*entity.User),
	}
}

func (r *InMemoryUserRepository) Create(ctx context.Context, user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if strings.EqualFold(u.Email, user.Email) {
			return appErrors.NewConflictError("email already registered")
		}
	}

	userCopy := *user
	r.users[user.ID] = &userCopy
	return nil
}

func (r *InMemoryUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if strings.EqualFold(u.Email, email) {
			userCopy := *u
			return &userCopy, nil
		}
	}

	return nil, appErrors.NewNotFoundError("user not found")
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.users[id]
	if !exists {
		return nil, appErrors.NewNotFoundError("user not found")
	}

	userCopy := *u
	return &userCopy, nil
}

func (r *InMemoryUserRepository) Update(ctx context.Context, user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return appErrors.NewNotFoundError("user not found for update")
	}

	userCopy := *user
	r.users[user.ID] = &userCopy
	return nil
}

func (r *InMemoryUserRepository) UpdateLastLoginAt(ctx context.Context, userID string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.users[userID]
	if !exists {
		return appErrors.NewNotFoundError("user not found for last_login_at update")
	}

	utcTime := at.UTC()
	u.LastLoginAt = &utcTime
	u.UpdatedAt = utcTime
	return nil
}

func (r *InMemoryUserRepository) UpdateLastActivityAt(ctx context.Context, userID string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.users[userID]
	if !exists {
		return appErrors.NewNotFoundError("user not found for last_activity_at update")
	}

	utcTime := at.UTC()
	u.LastActivityAt = &utcTime
	u.UpdatedAt = utcTime
	return nil
}

// InMemoryTokenRepository is a thread-safe in-memory implementation of TokenRepository.
type InMemoryTokenRepository struct {
	mu     sync.RWMutex
	tokens map[string]tokenEntry // key: tokenID
}

type tokenEntry struct {
	userID    string
	expiresAt time.Time
}

func NewInMemoryTokenRepository() repository.TokenRepository {
	return &InMemoryTokenRepository{
		tokens: make(map[string]tokenEntry),
	}
}

func (r *InMemoryTokenRepository) StoreRefreshToken(ctx context.Context, tokenID string, userID string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tokens[tokenID] = tokenEntry{
		userID:    userID,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *InMemoryTokenRepository) ValidateRefreshToken(ctx context.Context, tokenID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.tokens[tokenID]
	if !exists {
		return "", appErrors.NewUnauthorizedError("refresh token has been revoked or expired")
	}

	if time.Now().After(entry.expiresAt) {
		return "", appErrors.NewUnauthorizedError("refresh token has expired")
	}

	return entry.userID, nil
}

func (r *InMemoryTokenRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.tokens, tokenID)
	return nil
}

func (r *InMemoryTokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for tid, entry := range r.tokens {
		if entry.userID == userID {
			delete(r.tokens, tid)
		}
	}
	return nil
}
