// Package user_service implements the use-cases for user management and authentication.
package user_service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
	"github.com/truemycornea/aura/backend/internal/domain/user"
	"github.com/truemycornea/aura/backend/pkg/config"
)

// TokenPair holds the access and refresh JWT strings issued after authentication.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// claims is the JWT payload.
type claims struct {
	UserID string    `json:"user_id"`
	Email  string    `json:"email"`
	Role   user.Role `json:"role"`
	jwt.RegisteredClaims
}

// Service implements all user-related use cases.
type Service struct {
	repo   user.Repository
	cfg    config.AuthConfig
	logger *zap.Logger
}

// New creates a new user Service.
func New(repo user.Repository, cfg config.AuthConfig, logger *zap.Logger) *Service {
	return &Service{repo: repo, cfg: cfg, logger: logger}
}

// Register creates a new local user account.
func (s *Service) Register(ctx context.Context, email, password, displayName string) (*user.User, error) {
	existing, _ := s.repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, apierrors.ErrAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	u := &user.User{
		ID:           uuid.New(),
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: string(hash),
		Role:         user.RoleUser,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	s.logger.Info("user registered", zap.String("email", email), zap.String("id", u.ID.String()))
	return u, nil
}

// Login verifies credentials and returns a JWT token pair.
func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, apierrors.ErrUnauthorized
	}

	if !u.IsActive {
		return nil, apierrors.ErrForbidden
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, apierrors.ErrUnauthorized
	}

	tokens, err := s.issueTokens(u)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	u.LastLoginAt = &now
	_ = s.repo.Update(ctx, u)

	return tokens, nil
}

// GetByID returns a user by primary key.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.repo.GetByID(ctx, id)
}

// ListUsers returns a paginated list of users (admin only — callers must enforce RBAC).
func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]*user.User, int64, error) {
	return s.repo.List(ctx, limit, offset)
}

// ValidateToken parses a JWT and returns the userID, role, and any error.
// This signature satisfies the middleware.TokenValidator interface.
func (s *Service) ValidateToken(tokenStr string) (string, string, error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", "", apierrors.ErrUnauthorized
	}
	return c.UserID, string(c.Role), nil
}

func (s *Service) issueTokens(u *user.User) (*TokenPair, error) {
	now := time.Now()
	exp := now.Add(s.cfg.AccessTokenExpiry)

	accessClaims := &claims{
		UserID: u.ID.String(),
		Email:  u.Email,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   u.ID.String(),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	refreshClaims := &claims{
		UserID: u.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   u.ID.String(),
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("signing refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    exp,
	}, nil
}
