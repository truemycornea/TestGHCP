// Package album defines the Album domain entity and sharing model.
package album

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SharePermission controls what a share-link recipient may do.
type SharePermission string

const (
	PermissionViewOnly   SharePermission = "view_only"
	PermissionContributor SharePermission = "contributor"
)

// SharedLink represents a public share link for an album.
type SharedLink struct {
	ID          uuid.UUID       `json:"id"`
	AlbumID     uuid.UUID       `json:"album_id"`
	Token       string          `json:"token"` // random URL-safe token
	Permission  SharePermission `json:"permission"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty"`
	PasswordHash string         `json:"-"` // optional password protection
	CreatedAt   time.Time       `json:"created_at"`
}

// Album groups assets and may be shared with other users.
type Album struct {
	ID          uuid.UUID  `json:"id"`
	OwnerID     uuid.UUID  `json:"owner_id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	CoverAssetID *uuid.UUID `json:"cover_asset_id,omitempty"`
	AssetCount  int        `json:"asset_count"`
	SharedLinks []SharedLink `json:"shared_links,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Repository defines the persistence port for Album entities.
type Repository interface {
	Create(ctx context.Context, a *Album) error
	GetByID(ctx context.Context, id uuid.UUID) (*Album, error)
	Update(ctx context.Context, a *Album) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Album, int64, error)
	AddAsset(ctx context.Context, albumID, assetID uuid.UUID) error
	RemoveAsset(ctx context.Context, albumID, assetID uuid.UUID) error
	ListAssets(ctx context.Context, albumID uuid.UUID, limit, offset int) ([]*uuid.UUID, error)
	CreateSharedLink(ctx context.Context, link *SharedLink) error
	GetSharedLinkByToken(ctx context.Context, token string) (*SharedLink, error)
	DeleteSharedLink(ctx context.Context, id uuid.UUID) error
}
