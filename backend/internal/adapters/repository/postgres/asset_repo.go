// Package postgres provides PostgreSQL implementations of all domain repository interfaces.
package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/truemycornea/aura/backend/internal/domain/asset"
	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
)

// AssetRepository is the PostgreSQL implementation of asset.Repository.
type AssetRepository struct {
	pool *pgxpool.Pool
}

// NewAssetRepository creates a new AssetRepository backed by a pgxpool.
func NewAssetRepository(pool *pgxpool.Pool) *AssetRepository {
	return &AssetRepository{pool: pool}
}

func (r *AssetRepository) Create(ctx context.Context, a *asset.Asset) error {
	exifJSON, err := json.Marshal(a.EXIF)
	if err != nil {
		return fmt.Errorf("marshalling EXIF: %w", err)
	}

	const q = `
		INSERT INTO assets (
			id, owner_id, filename, original_path, thumb_path, proxy_path,
			mime_type, media_type, file_size, duration, checksum, phash_value,
			exif_data, is_favourite, is_archived, is_trashed, is_processed,
			taken_at, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
		)`

	_, err = r.pool.Exec(ctx, q,
		a.ID, a.OwnerID, a.Filename, a.OriginalPath, a.ThumbPath, a.ProxyPath,
		a.MimeType, string(a.MediaType), a.FileSize, a.Duration, a.Checksum, a.PHashValue,
		exifJSON, a.IsFavourite, a.IsArchived, a.IsTrashed, a.IsProcessed,
		a.TakenAt, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *AssetRepository) GetByID(ctx context.Context, id uuid.UUID) (*asset.Asset, error) {
	const q = `SELECT * FROM assets WHERE id = $1 AND is_trashed = false`
	row := r.pool.QueryRow(ctx, q, id)
	a, err := scanAsset(row)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	return a, nil
}

func (r *AssetRepository) GetByChecksum(ctx context.Context, checksum string) (*asset.Asset, error) {
	const q = `SELECT * FROM assets WHERE checksum = $1 LIMIT 1`
	row := r.pool.QueryRow(ctx, q, checksum)
	a, err := scanAsset(row)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	return a, nil
}

func (r *AssetRepository) GetByPHash(ctx context.Context, phash string) ([]*asset.Asset, error) {
	const q = `SELECT * FROM assets WHERE phash_value = $1`
	rows, err := r.pool.Query(ctx, q, phash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectAssets(rows)
}

func (r *AssetRepository) Update(ctx context.Context, a *asset.Asset) error {
	exifJSON, err := json.Marshal(a.EXIF)
	if err != nil {
		return fmt.Errorf("marshalling EXIF: %w", err)
	}
	a.UpdatedAt = time.Now()

	const q = `
		UPDATE assets SET
			filename=$2, thumb_path=$3, proxy_path=$4, phash_value=$5, exif_data=$6,
			is_favourite=$7, is_archived=$8, is_trashed=$9, is_processed=$10,
			taken_at=$11, updated_at=$12, clip_embedding=$13
		WHERE id=$1`

	var embeddingStr any
	if len(a.CLIPEmbedding) > 0 {
		embeddingStr = vectorLiteral(a.CLIPEmbedding)
	}

	_, err = r.pool.Exec(ctx, q,
		a.ID, a.Filename, a.ThumbPath, a.ProxyPath, a.PHashValue, exifJSON,
		a.IsFavourite, a.IsArchived, a.IsTrashed, a.IsProcessed,
		a.TakenAt, a.UpdatedAt, embeddingStr,
	)
	return err
}

func (r *AssetRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE assets SET is_trashed=true, updated_at=$2 WHERE id=$1`, id, time.Now())
	return err
}

func (r *AssetRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM assets WHERE id=$1`, id)
	return err
}

func (r *AssetRepository) List(ctx context.Context, filter asset.ListFilter, limit, offset int) ([]*asset.Asset, int64, error) {
	conditions := []string{"1=1"}
	args := []any{}
	i := 1

	if filter.OwnerID != nil {
		conditions = append(conditions, fmt.Sprintf("owner_id=$%d", i))
		args = append(args, *filter.OwnerID)
		i++
	}
	if filter.MediaType != nil {
		conditions = append(conditions, fmt.Sprintf("media_type=$%d", i))
		args = append(args, string(*filter.MediaType))
		i++
	}
	if filter.IsFav != nil {
		conditions = append(conditions, fmt.Sprintf("is_favourite=$%d", i))
		args = append(args, *filter.IsFav)
		i++
	}
	if filter.IsArchived != nil {
		conditions = append(conditions, fmt.Sprintf("is_archived=$%d", i))
		args = append(args, *filter.IsArchived)
		i++
	}
	if filter.IsTrashed != nil {
		conditions = append(conditions, fmt.Sprintf("is_trashed=$%d", i))
		args = append(args, *filter.IsTrashed)
		i++
	} else {
		conditions = append(conditions, "is_trashed=false")
	}
	if filter.City != "" {
		conditions = append(conditions, fmt.Sprintf("exif_data->>'city'=$%d", i))
		args = append(args, filter.City)
		i++
	}
	if filter.Country != "" {
		conditions = append(conditions, fmt.Sprintf("exif_data->>'country'=$%d", i))
		args = append(args, filter.Country)
		i++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("taken_at >= $%d", i))
		args = append(args, *filter.DateFrom)
		i++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("taken_at <= $%d", i))
		args = append(args, *filter.DateTo)
		i++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM assets WHERE %s`, where)
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append(args, limit, offset)
	listQ := fmt.Sprintf(`SELECT * FROM assets WHERE %s ORDER BY taken_at DESC LIMIT $%d OFFSET $%d`, where, i, i+1)

	rows, err := r.pool.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	assets, err := collectAssets(rows)
	return assets, total, err
}

func (r *AssetRepository) SemanticSearch(ctx context.Context, queryVec []float32, limit int) ([]*asset.Asset, error) {
	const q = `
		SELECT * FROM assets
		WHERE clip_embedding IS NOT NULL AND is_trashed = false
		ORDER BY clip_embedding <=> $1
		LIMIT $2`

	rows, err := r.pool.Query(ctx, q, vectorLiteral(queryVec), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectAssets(rows)
}

// vectorLiteral converts a float32 slice to a pgvector-compatible string literal.
func vectorLiteral(v []float32) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// pgxRow is satisfied by both pgx.Row and pgx.Rows to share scan logic.
type pgxRow interface {
	Scan(dest ...any) error
}

func scanAsset(row pgxRow) (*asset.Asset, error) {
	var a asset.Asset
	var exifJSON []byte
	var mediaTypeStr string

	err := row.Scan(
		&a.ID, &a.OwnerID, &a.Filename, &a.OriginalPath, &a.ThumbPath, &a.ProxyPath,
		&a.MimeType, &mediaTypeStr, &a.FileSize, &a.Duration, &a.Checksum, &a.PHashValue,
		&exifJSON, &a.IsFavourite, &a.IsArchived, &a.IsTrashed, &a.IsProcessed,
		&a.TakenAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	a.MediaType = asset.MediaType(mediaTypeStr)
	if len(exifJSON) > 0 {
		_ = json.Unmarshal(exifJSON, &a.EXIF)
	}
	return &a, nil
}

type pgxRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func collectAssets(rows pgxRows) ([]*asset.Asset, error) {
	var result []*asset.Asset
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}
