// Package album_service implements the use-cases for album management and sharing.
package album_service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/truemycornea/aura/backend/internal/domain/album"
	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
)

// Service implements album-related use cases.
type Service struct {
	repo   album.Repository
	logger *zap.Logger
}

// New creates a new album Service.
func New(repo album.Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// CreateAlbum creates a new album owned by the given user.
func (s *Service) CreateAlbum(ctx context.Context, ownerID uuid.UUID, name, description string) (*album.Album, error) {
	a := &album.Album{
		ID:          uuid.New(),
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("creating album: %w", err)
	}
	return a, nil
}

// GetAlbum retrieves an album, enforcing ownership.
func (s *Service) GetAlbum(ctx context.Context, id, ownerID uuid.UUID) (*album.Album, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	if a.OwnerID != ownerID {
		return nil, apierrors.ErrForbidden
	}
	return a, nil
}

// ListAlbums returns a paginated list of albums for an owner.
func (s *Service) ListAlbums(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*album.Album, int64, error) {
	return s.repo.List(ctx, ownerID, limit, offset)
}

// AddAssetToAlbum links an asset to an album.
func (s *Service) AddAssetToAlbum(ctx context.Context, albumID, assetID, ownerID uuid.UUID) error {
	a, err := s.repo.GetByID(ctx, albumID)
	if err != nil {
		return apierrors.ErrNotFound
	}
	if a.OwnerID != ownerID {
		return apierrors.ErrForbidden
	}
	return s.repo.AddAsset(ctx, albumID, assetID)
}

// CreateShareLink generates a public share link with optional password and expiry.
func (s *Service) CreateShareLink(ctx context.Context, albumID, ownerID uuid.UUID, perm album.SharePermission, password string, expiresAt *time.Time) (*album.SharedLink, error) {
	a, err := s.repo.GetByID(ctx, albumID)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	if a.OwnerID != ownerID {
		return nil, apierrors.ErrForbidden
	}

	token, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("generating share token: %w", err)
	}

	link := &album.SharedLink{
		ID:         uuid.New(),
		AlbumID:    albumID,
		Token:      token,
		Permission: perm,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}

	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hashing link password: %w", err)
		}
		link.PasswordHash = string(hash)
	}

	if err := s.repo.CreateSharedLink(ctx, link); err != nil {
		return nil, fmt.Errorf("persisting share link: %w", err)
	}

	s.logger.Info("share link created", zap.String("album_id", albumID.String()), zap.String("token", token))
	return link, nil
}

func generateToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
