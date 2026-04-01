// Package face defines the Face domain entity used for face clustering.
package face

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Person is a cluster of faces that the system has grouped together.
type Person struct {
	ID          uuid.UUID `json:"id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Name        string    `json:"name"`
	CoverFaceID *uuid.UUID `json:"cover_face_id,omitempty"`
	FaceCount   int       `json:"face_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Face is a single detected face region within an asset.
type Face struct {
	ID        uuid.UUID `json:"id"`
	AssetID   uuid.UUID `json:"asset_id"`
	PersonID  *uuid.UUID `json:"person_id,omitempty"` // nil until assigned to a Person
	// BoundingBox in normalised coordinates [0,1].
	BBoxX     float64   `json:"bbox_x"`
	BBoxY     float64   `json:"bbox_y"`
	BBoxW     float64   `json:"bbox_w"`
	BBoxH     float64   `json:"bbox_h"`
	// Embedding is a 512-dim facial embedding vector stored in pgvector.
	Embedding []float32 `json:"-"`
	Score     float64   `json:"score"` // detection confidence
	CreatedAt time.Time `json:"created_at"`
}

// Repository defines the persistence port for Face and Person entities.
type Repository interface {
	CreateFace(ctx context.Context, f *Face) error
	GetFaceByID(ctx context.Context, id uuid.UUID) (*Face, error)
	ListFacesByAsset(ctx context.Context, assetID uuid.UUID) ([]*Face, error)
	AssignFaceToPerson(ctx context.Context, faceID, personID uuid.UUID) error
	// FindSimilarFaces returns faces whose embedding is within cosine distance
	// of the query, ordered by similarity — used for clustering.
	FindSimilarFaces(ctx context.Context, embedding []float32, threshold float64, limit int) ([]*Face, error)

	CreatePerson(ctx context.Context, p *Person) error
	GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error)
	UpdatePerson(ctx context.Context, p *Person) error
	ListPersons(ctx context.Context, ownerID uuid.UUID) ([]*Person, error)
}
