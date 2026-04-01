-- =============================================================================
-- Aura – Initial Database Schema
-- Requires: PostgreSQL 15+ with the pgvector extension
-- =============================================================================

-- Enable the pgvector extension for storing CLIP and face embeddings.
CREATE EXTENSION IF NOT EXISTS vector;
-- Enable the uuid-ossp extension for UUID generation helpers.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- Enable pg_trgm for fast ILIKE / full-text style queries on tags and filenames.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- =============================================================================
-- Users
-- Stores local and OIDC-federated accounts with RBAC role assignment.
-- =============================================================================
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL,
    avatar_url    TEXT,
    role          TEXT NOT NULL DEFAULT 'user'
                    CHECK (role IN ('admin', 'user', 'viewer', 'contributor')),
    password_hash TEXT,           -- NULL for OIDC-only accounts
    oidc_subject  TEXT UNIQUE,    -- sub claim from the identity provider
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email        ON users (email);
CREATE INDEX idx_users_oidc_subject ON users (oidc_subject) WHERE oidc_subject IS NOT NULL;

-- =============================================================================
-- Assets
-- The central table for all photos and videos.  Original files are never
-- mutated (non-destructive pipeline).  EXIF and geocoding data are stored as
-- JSONB for schema flexibility.
-- =============================================================================
CREATE TABLE IF NOT EXISTS assets (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Storage references (S3-compatible object keys)
    filename       TEXT NOT NULL,
    original_path  TEXT NOT NULL,                    -- key of the original file
    thumb_path     TEXT,                             -- key of the WebP thumbnail
    proxy_path     TEXT,                             -- key of the H.265/AV1 proxy (videos)

    -- File metadata
    mime_type      TEXT NOT NULL,
    media_type     TEXT NOT NULL CHECK (media_type IN ('photo', 'video')),
    file_size      BIGINT NOT NULL DEFAULT 0,
    duration       FLOAT,                            -- video duration in seconds
    checksum       TEXT NOT NULL,                    -- SHA-256 of the original file
    phash_value    TEXT,                             -- perceptual hash for deduplication

    -- Structured EXIF / XMP / IPTC data (flexible schema)
    exif_data      JSONB NOT NULL DEFAULT '{}',

    -- User-controlled flags
    is_favourite   BOOLEAN NOT NULL DEFAULT FALSE,
    is_archived    BOOLEAN NOT NULL DEFAULT FALSE,
    is_trashed     BOOLEAN NOT NULL DEFAULT FALSE,
    is_processed   BOOLEAN NOT NULL DEFAULT FALSE,   -- set after the worker pipeline completes

    -- Temporal
    taken_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- CLIP embedding for semantic / natural-language search (512-dim vector).
    -- Populated by the AI worker after upload.
    clip_embedding vector(512)
);

-- Standard access patterns
CREATE INDEX idx_assets_owner_id     ON assets (owner_id);
CREATE INDEX idx_assets_taken_at     ON assets (taken_at DESC);
CREATE INDEX idx_assets_media_type   ON assets (media_type);
CREATE INDEX idx_assets_is_trashed   ON assets (is_trashed);
CREATE INDEX idx_assets_checksum     ON assets (checksum);
CREATE INDEX idx_assets_phash        ON assets (phash_value) WHERE phash_value IS NOT NULL;

-- JSONB index for geospatial / city / country filtering
CREATE INDEX idx_assets_exif_city    ON assets USING GIN ((exif_data -> 'city')    gin_trgm_ops);
CREATE INDEX idx_assets_exif_country ON assets USING GIN ((exif_data -> 'country') gin_trgm_ops);
CREATE INDEX idx_assets_exif_data    ON assets USING GIN (exif_data);

-- pgvector IVFFlat index for approximate nearest-neighbour semantic search.
-- cosine distance (<=>): best for normalised CLIP embeddings.
-- lists=100 is a good starting point for up to ~1M rows; tune as the library grows.
CREATE INDEX idx_assets_clip_embedding
    ON assets USING ivfflat (clip_embedding vector_cosine_ops)
    WITH (lists = 100)
    WHERE clip_embedding IS NOT NULL;

-- =============================================================================
-- Tags
-- User-defined and AI-generated keyword labels.
-- =============================================================================
CREATE TABLE IF NOT EXISTS tags (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    source     TEXT NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'ai')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_id, name)
);

CREATE INDEX idx_tags_owner ON tags (owner_id);
CREATE INDEX idx_tags_name  ON tags USING GIN (name gin_trgm_ops);

-- Many-to-many: assets ↔ tags
CREATE TABLE IF NOT EXISTS asset_tags (
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    tag_id   UUID NOT NULL REFERENCES tags(id)   ON DELETE CASCADE,
    PRIMARY KEY (asset_id, tag_id)
);

CREATE INDEX idx_asset_tags_tag_id ON asset_tags (tag_id);

-- =============================================================================
-- Albums
-- Named collections that can be shared via expirable, optionally-password-
-- protected links.
-- =============================================================================
CREATE TABLE IF NOT EXISTS albums (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    description    TEXT,
    cover_asset_id UUID REFERENCES assets(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_albums_owner ON albums (owner_id);

-- Many-to-many: albums ↔ assets
CREATE TABLE IF NOT EXISTS album_assets (
    album_id UUID NOT NULL REFERENCES albums(id)  ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id)  ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (album_id, asset_id)
);

CREATE INDEX idx_album_assets_asset ON album_assets (asset_id);

-- =============================================================================
-- Shared Links
-- Each link is scoped to one album and carries its own expiry and permission.
-- =============================================================================
CREATE TABLE IF NOT EXISTS shared_links (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    album_id      UUID NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    token         TEXT NOT NULL UNIQUE,          -- URL-safe random token
    permission    TEXT NOT NULL DEFAULT 'view_only'
                    CHECK (permission IN ('view_only', 'contributor')),
    expires_at    TIMESTAMPTZ,                   -- NULL = never expires
    password_hash TEXT,                          -- NULL = no password
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shared_links_token    ON shared_links (token);
CREATE INDEX idx_shared_links_album_id ON shared_links (album_id);

-- =============================================================================
-- Persons (Face Clustering)
-- A Person is a named cluster of Face records belonging to one owner.
-- =============================================================================
CREATE TABLE IF NOT EXISTS persons (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          TEXT NOT NULL DEFAULT 'Unknown',
    cover_face_id UUID,                          -- filled in after FK is available
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_persons_owner ON persons (owner_id);

-- =============================================================================
-- Faces
-- Individual face detections within assets.  Each face holds a 512-dim
-- facial embedding for clustering and recognition.
-- =============================================================================
CREATE TABLE IF NOT EXISTS faces (
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    asset_id  UUID NOT NULL REFERENCES assets(id)  ON DELETE CASCADE,
    person_id UUID          REFERENCES persons(id) ON DELETE SET NULL,
    -- Bounding box in normalised [0,1] coordinates.
    bbox_x    FLOAT NOT NULL,
    bbox_y    FLOAT NOT NULL,
    bbox_w    FLOAT NOT NULL,
    bbox_h    FLOAT NOT NULL,
    -- 512-dim facial embedding (e.g. from InsightFace / DeepFace).
    embedding vector(512),
    score     FLOAT NOT NULL DEFAULT 0,          -- detection confidence
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_faces_asset_id  ON faces (asset_id);
CREATE INDEX idx_faces_person_id ON faces (person_id) WHERE person_id IS NOT NULL;

-- pgvector HNSW index for fast face-similarity search during clustering.
-- HNSW provides better recall than IVFFlat for small-to-medium datasets.
CREATE INDEX idx_faces_embedding
    ON faces USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64)
    WHERE embedding IS NOT NULL;

-- Add the deferred FK from persons → faces once the faces table exists.
ALTER TABLE persons
    ADD CONSTRAINT fk_persons_cover_face
    FOREIGN KEY (cover_face_id) REFERENCES faces(id) ON DELETE SET NULL;

-- =============================================================================
-- Background Job Audit Log (optional — useful for observability)
-- =============================================================================
CREATE TABLE IF NOT EXISTS job_log (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    asset_id   UUID REFERENCES assets(id) ON DELETE SET NULL,
    job_type   TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    error_msg  TEXT,
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_job_log_asset_id ON job_log (asset_id);
CREATE INDEX idx_job_log_status   ON job_log (status);
