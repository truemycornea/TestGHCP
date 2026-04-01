// Package user defines the User domain entity and its repository interface.
package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Role represents the RBAC role of a user within the system.
type Role string

const (
	RoleAdmin       Role = "admin"
	RoleUser        Role = "user"
	RoleViewer      Role = "viewer"
	RoleContributor Role = "contributor"
)

// User is the core domain entity representing a registered account.
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	AvatarURL    string     `json:"avatar_url,omitempty"`
	Role         Role       `json:"role"`
	PasswordHash string     `json:"-"` // never serialised
	OIDCSubject  string     `json:"-"` // external identity provider subject
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// Repository defines the persistence port for User entities.
// Implementations live in the adapters layer.
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByOIDCSubject(ctx context.Context, subject string) (*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*User, int64, error)
}
