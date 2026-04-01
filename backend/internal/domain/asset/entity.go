// Package asset defines the Asset domain entity and its repository interface.
// An Asset can be either a photo or a video stored in object storage.
package asset

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MediaType classifies an asset as either a photo or a video.
type MediaType string

const (
	MediaTypePhoto MediaType = "photo"
	MediaTypeVideo MediaType = "video"
)

// EXIFData holds the raw metadata extracted from an asset file.
type EXIFData struct {
	Make             string    `json:"make,omitempty"`
	Model            string    `json:"model,omitempty"`
	LensModel        string    `json:"lens_model,omitempty"`
	FNumber          float64   `json:"f_number,omitempty"`
	ExposureTime     string    `json:"exposure_time,omitempty"`
	ISO              int       `json:"iso,omitempty"`
	FocalLength      float64   `json:"focal_length,omitempty"`
	Width            int       `json:"width,omitempty"`
	Height           int       `json:"height,omitempty"`
	Orientation      int       `json:"orientation,omitempty"`
	DateTimeOriginal time.Time `json:"date_time_original,omitempty"`
	GPSLatitude      float64   `json:"gps_latitude,omitempty"`
	GPSLongitude     float64   `json:"gps_longitude,omitempty"`
	GPSAltitude      float64   `json:"gps_altitude,omitempty"`
	City             string    `json:"city,omitempty"`
	Country          string    `json:"country,omitempty"`
}

// Asset is the core media entity managed by the platform.
type Asset struct {
	ID           uuid.UUID `json:"id"`
	OwnerID      uuid.UUID `json:"owner_id"`
	Filename     string    `json:"filename"`
	OriginalPath string    `json:"original_path"` // S3 key of the original file
	ThumbPath    string    `json:"thumb_path"`    // S3 key of the WebP thumbnail
	ProxyPath    string    `json:"proxy_path"`    // S3 key for video proxy (H.265/AV1)
	MimeType     string    `json:"mime_type"`
	MediaType    MediaType `json:"media_type"`
	FileSize     int64     `json:"file_size"`
	Duration     float64   `json:"duration,omitempty"` // video duration in seconds
	Checksum     string    `json:"checksum"`           // SHA-256 of original file
	PHashValue   string    `json:"phash_value"`        // perceptual hash for deduplication
	EXIF         EXIFData  `json:"exif"`
	// CLIPEmbedding is a 512-dim vector stored in pgvector for semantic search.
	CLIPEmbedding []float32 `json:"-"`
	IsFavourite   bool      `json:"is_favourite"`
	IsArchived    bool      `json:"is_archived"`
	IsTrashed     bool      `json:"is_trashed"`
	IsProcessed   bool      `json:"is_processed"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TakenAt       time.Time `json:"taken_at"`
}

// ListFilter carries optional predicates for listing assets.
type ListFilter struct {
	OwnerID    *uuid.UUID
	MediaType  *MediaType
	IsFav      *bool
	IsArchived *bool
	IsTrashed  *bool
	City       string
	Country    string
	DateFrom   *time.Time
	DateTo     *time.Time
}

// Repository defines the persistence port for Asset entities.
type Repository interface {
	Create(ctx context.Context, a *Asset) error
	GetByID(ctx context.Context, id uuid.UUID) (*Asset, error)
	GetByChecksum(ctx context.Context, checksum string) (*Asset, error)
	GetByPHash(ctx context.Context, phash string) ([]*Asset, error)
	Update(ctx context.Context, a *Asset) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListFilter, limit, offset int) ([]*Asset, int64, error)
	// SemanticSearch returns assets whose CLIP embedding is within the given
	// cosine distance of the query vector, ordered by similarity.
	SemanticSearch(ctx context.Context, queryVec []float32, limit int) ([]*Asset, error)
}
