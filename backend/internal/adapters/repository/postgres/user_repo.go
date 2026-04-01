package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/truemycornea/aura/backend/internal/domain/user"
	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
)

// UserRepository is the PostgreSQL implementation of user.Repository.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	const q = `
		INSERT INTO users (id, email, display_name, avatar_url, role, password_hash, oidc_subject, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.pool.Exec(ctx, q,
		u.ID, u.Email, u.DisplayName, u.AvatarURL, string(u.Role),
		u.PasswordHash, u.OIDCSubject, u.IsActive, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT * FROM users WHERE id=$1 AND is_active=true`, id)
	return scanUser(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT * FROM users WHERE email=$1`, email)
	u, err := scanUser(row)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	return u, nil
}

func (r *UserRepository) GetByOIDCSubject(ctx context.Context, subject string) (*user.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT * FROM users WHERE oidc_subject=$1`, subject)
	u, err := scanUser(row)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	return u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	u.UpdatedAt = time.Now()
	const q = `
		UPDATE users SET
			display_name=$2, avatar_url=$3, role=$4, is_active=$5,
			last_login_at=$6, updated_at=$7
		WHERE id=$1`
	_, err := r.pool.Exec(ctx, q,
		u.ID, u.DisplayName, u.AvatarURL, string(u.Role),
		u.IsActive, u.LastLoginAt, u.UpdatedAt,
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET is_active=false WHERE id=$1`, id)
	return err
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning user row: %w", err)
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func scanUser(row pgxRow) (*user.User, error) {
	var u user.User
	var roleStr string
	err := row.Scan(
		&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &roleStr,
		&u.PasswordHash, &u.OIDCSubject, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	if err != nil {
		return nil, err
	}
	u.Role = user.Role(roleStr)
	return &u, nil
}
