package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/truemycornea/aura/backend/internal/domain/album"
	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
)

// AlbumRepository is the PostgreSQL implementation of album.Repository.
type AlbumRepository struct {
	pool *pgxpool.Pool
}

// NewAlbumRepository creates a new AlbumRepository.
func NewAlbumRepository(pool *pgxpool.Pool) *AlbumRepository {
	return &AlbumRepository{pool: pool}
}

func (r *AlbumRepository) Create(ctx context.Context, a *album.Album) error {
	const q = `INSERT INTO albums (id, owner_id, name, description, cover_asset_id, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.pool.Exec(ctx, q, a.ID, a.OwnerID, a.Name, a.Description, a.CoverAssetID, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r *AlbumRepository) GetByID(ctx context.Context, id uuid.UUID) (*album.Album, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, owner_id, name, description, cover_asset_id, created_at, updated_at,
		 (SELECT COUNT(*) FROM album_assets WHERE album_id=albums.id) AS asset_count
		 FROM albums WHERE id=$1`, id)
	return scanAlbum(row)
}

func (r *AlbumRepository) Update(ctx context.Context, a *album.Album) error {
	a.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE albums SET name=$2, description=$3, cover_asset_id=$4, updated_at=$5 WHERE id=$1`,
		a.ID, a.Name, a.Description, a.CoverAssetID, a.UpdatedAt)
	return err
}

func (r *AlbumRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM albums WHERE id=$1`, id)
	return err
}

func (r *AlbumRepository) List(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*album.Album, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM albums WHERE owner_id=$1`, ownerID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_id, name, description, cover_asset_id, created_at, updated_at,
		 (SELECT COUNT(*) FROM album_assets WHERE album_id=albums.id) AS asset_count
		 FROM albums WHERE owner_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var albums []*album.Album
	for rows.Next() {
		a, err := scanAlbum(rows)
		if err != nil {
			return nil, 0, err
		}
		albums = append(albums, a)
	}
	return albums, total, rows.Err()
}

func (r *AlbumRepository) AddAsset(ctx context.Context, albumID, assetID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO album_assets (album_id, asset_id, added_at) VALUES ($1,$2,NOW()) ON CONFLICT DO NOTHING`,
		albumID, assetID)
	return err
}

func (r *AlbumRepository) RemoveAsset(ctx context.Context, albumID, assetID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM album_assets WHERE album_id=$1 AND asset_id=$2`, albumID, assetID)
	return err
}

func (r *AlbumRepository) ListAssets(ctx context.Context, albumID uuid.UUID, limit, offset int) ([]*uuid.UUID, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT asset_id FROM album_assets WHERE album_id=$1 ORDER BY added_at DESC LIMIT $2 OFFSET $3`,
		albumID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []*uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, &id)
	}
	return ids, rows.Err()
}

func (r *AlbumRepository) CreateSharedLink(ctx context.Context, link *album.SharedLink) error {
	const q = `INSERT INTO shared_links (id, album_id, token, permission, expires_at, password_hash, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.pool.Exec(ctx, q, link.ID, link.AlbumID, link.Token, string(link.Permission), link.ExpiresAt, link.PasswordHash, link.CreatedAt)
	return err
}

func (r *AlbumRepository) GetSharedLinkByToken(ctx context.Context, token string) (*album.SharedLink, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, album_id, token, permission, expires_at, password_hash, created_at FROM shared_links WHERE token=$1`, token)
	var l album.SharedLink
	var permStr string
	err := row.Scan(&l.ID, &l.AlbumID, &l.Token, &permStr, &l.ExpiresAt, &l.PasswordHash, &l.CreatedAt)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	l.Permission = album.SharePermission(permStr)
	return &l, nil
}

func (r *AlbumRepository) DeleteSharedLink(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM shared_links WHERE id=$1`, id)
	return err
}

func scanAlbum(row pgxRow) (*album.Album, error) {
	var a album.Album
	err := row.Scan(&a.ID, &a.OwnerID, &a.Name, &a.Description, &a.CoverAssetID, &a.CreatedAt, &a.UpdatedAt, &a.AssetCount)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	return &a, nil
}
