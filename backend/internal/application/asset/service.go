// Package asset_service implements the use-cases for asset ingestion and management.
package asset_service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/truemycornea/aura/backend/internal/domain/asset"
	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
)

// StoragePort abstracts object-storage operations so the service remains
// independent of any specific S3 implementation.
type StoragePort interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	PresignGetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

// TaskQueuePort lets the asset service enqueue background jobs (transcoding,
// AI tagging, EXIF extraction, thumbnail generation).
type TaskQueuePort interface {
	EnqueueAssetProcessing(ctx context.Context, assetID uuid.UUID) error
}

// Service implements all asset-related use cases.
type Service struct {
	repo    asset.Repository
	storage StoragePort
	queue   TaskQueuePort
	logger  *zap.Logger
}

// New creates a new asset Service.
func New(repo asset.Repository, storage StoragePort, queue TaskQueuePort, logger *zap.Logger) *Service {
	return &Service{repo: repo, storage: storage, queue: queue, logger: logger}
}

// UploadAsset stores the original file, persists the asset record, and
// enqueues it for background processing (EXIF, thumbnail, AI tagging).
func (s *Service) UploadAsset(ctx context.Context, ownerID uuid.UUID, filename string, r io.Reader, size int64) (*asset.Asset, error) {
	// Detect media type from extension.
	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	mediaType := asset.MediaTypePhoto
	if strings.HasPrefix(mimeType, "video/") {
		mediaType = asset.MediaTypeVideo
	}

	// Compute SHA-256 while streaming to storage.
	pr, pw := io.Pipe()
	h := sha256.New()
	tr := io.TeeReader(r, h)

	errCh := make(chan error, 1)
	id := uuid.New()
	origKey := fmt.Sprintf("originals/%s/%s%s", ownerID, id, ext)

	go func() {
		defer pw.Close()
		_, err := io.Copy(pw, tr)
		errCh <- err
	}()

	if err := s.storage.Upload(ctx, origKey, pr, size, mimeType); err != nil {
		return nil, fmt.Errorf("%w: %s", apierrors.ErrStorageFailure, err)
	}
	if err := <-errCh; err != nil {
		return nil, fmt.Errorf("reading upload stream: %w", err)
	}

	checksum := fmt.Sprintf("%x", h.Sum(nil))

	// Reject exact duplicates.
	if existing, _ := s.repo.GetByChecksum(ctx, checksum); existing != nil {
		return nil, fmt.Errorf("%w: file already exists (id=%s)", apierrors.ErrAlreadyExists, existing.ID)
	}

	a := &asset.Asset{
		ID:           id,
		OwnerID:      ownerID,
		Filename:     filename,
		OriginalPath: origKey,
		MimeType:     mimeType,
		MediaType:    mediaType,
		FileSize:     size,
		Checksum:     checksum,
		IsProcessed:  false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("persisting asset: %w", err)
	}

	// Fire-and-forget background processing (EXIF, thumbnail, AI tagging).
	if err := s.queue.EnqueueAssetProcessing(ctx, a.ID); err != nil {
		s.logger.Warn("failed to enqueue asset processing", zap.String("asset_id", a.ID.String()), zap.Error(err))
	}

	s.logger.Info("asset uploaded", zap.String("id", a.ID.String()), zap.String("filename", filename))
	return a, nil
}

// GetAsset retrieves an asset and generates a pre-signed URL for the original file.
func (s *Service) GetAsset(ctx context.Context, id uuid.UUID) (*asset.Asset, string, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, "", apierrors.ErrNotFound
	}

	url, err := s.storage.PresignGetURL(ctx, a.OriginalPath, 1*time.Hour)
	if err != nil {
		return nil, "", fmt.Errorf("%w: presign URL: %s", apierrors.ErrStorageFailure, err)
	}

	return a, url, nil
}

// ListAssets returns a paginated, filtered list of assets owned by a user.
func (s *Service) ListAssets(ctx context.Context, ownerID uuid.UUID, filter asset.ListFilter, limit, offset int) ([]*asset.Asset, int64, error) {
	filter.OwnerID = &ownerID
	return s.repo.List(ctx, filter, limit, offset)
}

// TrashAsset moves an asset to the trash (soft-delete).
func (s *Service) TrashAsset(ctx context.Context, id, ownerID uuid.UUID) error {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return apierrors.ErrNotFound
	}
	if a.OwnerID != ownerID {
		return apierrors.ErrForbidden
	}
	a.IsTrashed = true
	a.UpdatedAt = time.Now()
	return s.repo.Update(ctx, a)
}

// FavouriteAsset toggles the favourite flag.
func (s *Service) FavouriteAsset(ctx context.Context, id, ownerID uuid.UUID, fav bool) error {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return apierrors.ErrNotFound
	}
	if a.OwnerID != ownerID {
		return apierrors.ErrForbidden
	}
	a.IsFavourite = fav
	a.UpdatedAt = time.Now()
	return s.repo.Update(ctx, a)
}
