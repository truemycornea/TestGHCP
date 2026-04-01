package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/truemycornea/aura/backend/internal/domain/face"
	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
)

// FaceRepository is the PostgreSQL implementation of face.Repository.
type FaceRepository struct {
	pool *pgxpool.Pool
}

// NewFaceRepository creates a new FaceRepository.
func NewFaceRepository(pool *pgxpool.Pool) *FaceRepository {
	return &FaceRepository{pool: pool}
}

func (r *FaceRepository) CreateFace(ctx context.Context, f *face.Face) error {
	const q = `
		INSERT INTO faces (id, asset_id, person_id, bbox_x, bbox_y, bbox_w, bbox_h, embedding, score, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.pool.Exec(ctx, q,
		f.ID, f.AssetID, f.PersonID, f.BBoxX, f.BBoxY, f.BBoxW, f.BBoxH,
		vectorLiteral(f.Embedding), f.Score, f.CreatedAt,
	)
	return err
}

func (r *FaceRepository) GetFaceByID(ctx context.Context, id uuid.UUID) (*face.Face, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, asset_id, person_id, bbox_x, bbox_y, bbox_w, bbox_h, score, created_at FROM faces WHERE id=$1`, id)
	return scanFace(row)
}

func (r *FaceRepository) ListFacesByAsset(ctx context.Context, assetID uuid.UUID) ([]*face.Face, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, asset_id, person_id, bbox_x, bbox_y, bbox_w, bbox_h, score, created_at FROM faces WHERE asset_id=$1`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectFaces(rows)
}

func (r *FaceRepository) AssignFaceToPerson(ctx context.Context, faceID, personID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE faces SET person_id=$2 WHERE id=$1`, faceID, personID)
	return err
}

func (r *FaceRepository) FindSimilarFaces(ctx context.Context, embedding []float32, threshold float64, limit int) ([]*face.Face, error) {
	const q = `
		SELECT id, asset_id, person_id, bbox_x, bbox_y, bbox_w, bbox_h, score, created_at
		FROM faces
		WHERE embedding <=> $1 < $2
		ORDER BY embedding <=> $1
		LIMIT $3`
	rows, err := r.pool.Query(ctx, q, vectorLiteral(embedding), threshold, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectFaces(rows)
}

func (r *FaceRepository) CreatePerson(ctx context.Context, p *face.Person) error {
	const q = `INSERT INTO persons (id, owner_id, name, cover_face_id, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := r.pool.Exec(ctx, q, p.ID, p.OwnerID, p.Name, p.CoverFaceID, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *FaceRepository) GetPersonByID(ctx context.Context, id uuid.UUID) (*face.Person, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, owner_id, name, cover_face_id, created_at, updated_at,
		 (SELECT COUNT(*) FROM faces WHERE person_id=persons.id) AS face_count
		 FROM persons WHERE id=$1`, id)
	return scanPerson(row)
}

func (r *FaceRepository) UpdatePerson(ctx context.Context, p *face.Person) error {
	p.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, `UPDATE persons SET name=$2, cover_face_id=$3, updated_at=$4 WHERE id=$1`,
		p.ID, p.Name, p.CoverFaceID, p.UpdatedAt)
	return err
}

func (r *FaceRepository) ListPersons(ctx context.Context, ownerID uuid.UUID) ([]*face.Person, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_id, name, cover_face_id, created_at, updated_at,
		 (SELECT COUNT(*) FROM faces WHERE person_id=persons.id) AS face_count
		 FROM persons WHERE owner_id=$1 ORDER BY name`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []*face.Person
	for rows.Next() {
		p, err := scanPerson(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning person: %w", err)
		}
		persons = append(persons, p)
	}
	return persons, rows.Err()
}

func scanFace(row pgxRow) (*face.Face, error) {
	var f face.Face
	err := row.Scan(&f.ID, &f.AssetID, &f.PersonID, &f.BBoxX, &f.BBoxY, &f.BBoxW, &f.BBoxH, &f.Score, &f.CreatedAt)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	return &f, nil
}

func collectFaces(rows pgxRows) ([]*face.Face, error) {
	var result []*face.Face
	for rows.Next() {
		f, err := scanFace(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, rows.Err()
}

func scanPerson(row pgxRow) (*face.Person, error) {
	var p face.Person
	err := row.Scan(&p.ID, &p.OwnerID, &p.Name, &p.CoverFaceID, &p.CreatedAt, &p.UpdatedAt, &p.FaceCount)
	if err != nil {
		return nil, apierrors.ErrNotFound
	}
	return &p, nil
}

// ensure vectorLiteral is accessible (defined in asset_repo.go in this package)
var _ = func() bool { return true }
