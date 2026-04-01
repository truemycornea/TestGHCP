// Package tag defines the Tag domain entity for asset tagging.
package tag

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Source indicates how a tag was applied.
type Source string

const (
	SourceManual Source = "manual"
	SourceAI     Source = "ai"
)

// Tag is a keyword or label associated with one or more assets.
type Tag struct {
	ID        uuid.UUID `json:"id"`
	OwnerID   uuid.UUID `json:"owner_id"`
	Name      string    `json:"name"`
	Source    Source    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository defines the persistence port for Tag entities.
type Repository interface {
	Create(ctx context.Context, t *Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tag, error)
	GetByName(ctx context.Context, ownerID uuid.UUID, name string) (*Tag, error)
	List(ctx context.Context, ownerID uuid.UUID) ([]*Tag, error)
	Delete(ctx context.Context, id uuid.UUID) error

	AddTagToAsset(ctx context.Context, tagID, assetID uuid.UUID) error
	RemoveTagFromAsset(ctx context.Context, tagID, assetID uuid.UUID) error
	ListTagsByAsset(ctx context.Context, assetID uuid.UUID) ([]*Tag, error)
}
