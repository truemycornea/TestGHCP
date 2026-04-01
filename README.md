# Aura – Next-Gen AI Image & Video Repository

> A high-performance, self-hostable media management platform inspired by Immich, Google Photos, and Synology Photos.

## ✨ Feature Highlights

| # | Feature | Status |
|---|---------|--------|
| 1 | **Enterprise Identity (IAM)** – OIDC/SAML + RBAC | 🏗 Boilerplate |
| 2 | **Hybrid AI Engine** – Local Ollama + OpenAI/Vertex AI | 🏗 Boilerplate |
| 3 | **Semantic Search** – CLIP embeddings via pgvector | 🏗 Boilerplate |
| 4 | **Smart Deduplication** – SHA-256 + pHash | 🏗 Boilerplate |
| 5 | **Contextual Editing Suggestions** – EXIF-driven AI | 🏗 Boilerplate |
| 6 | **Geospatial Intelligence** – GPS → city/landmark | 🏗 Boilerplate |
| 7 | **Multi-Tier Compression** – WebP thumbs + H.265 proxies | 🏗 Boilerplate |
| 8 | **Face Grouping & Recognition** | 🏗 Boilerplate |
| 9 | **Instant Sync & Background Upload** | �� Boilerplate |
| 10 | **Partner Sharing & Shared Albums** – expiry + password | 🏗 Boilerplate |

---

## 🏛 Architecture

```
aura/
├── backend/                  # Go (Gin) – Hexagonal / Clean Architecture
│   ├── cmd/api/              # Application entry point
│   ├── internal/
│   │   ├── domain/           # Entities & repository interfaces (ports)
│   │   │   ├── asset/
│   │   │   ├── user/
│   │   │   ├── album/
│   │   │   ├── face/
│   │   │   └── tag/
│   │   ├── application/      # Use-case services
│   │   │   ├── asset/
│   │   │   ├── user/
│   │   │   ├── album/
│   │   │   └── search/
│   │   └── adapters/         # Driven & driving adapters
│   │       ├── http/         # Gin handlers + middleware
│   │       ├── repository/   # PostgreSQL implementations
│   │       ├── storage/      # S3 / MinIO adapter
│   │       └── queue/        # Redis task queue
│   ├── pkg/                  # Shared utilities (config, logger, errors)
│   ├── migrations/           # SQL schema with pgvector indices
│   └── Dockerfile
│
├── frontend/                 # Next.js 14 (App Router) + Tailwind + Radix UI
│   ├── src/
│   │   ├── app/              # App Router pages
│   │   ├── components/       # React components (gallery, layout, ui)
│   │   └── lib/              # API client, hooks, TypeScript types
│   └── Dockerfile
│
├── docker-compose.yml        # Single-click deployment
└── .env.example              # Environment variable reference
```

### Technology Stack

| Layer | Technology | Reason |
|-------|-----------|--------|
| **Frontend** | Next.js 14 (App Router), Tailwind CSS, Radix UI | SSR performance; type-safe; accessible component primitives |
| **Backend** | Go (Gin), Hexagonal Architecture | Low memory, high concurrency, static binary |
| **Database** | PostgreSQL 16 + pgvector | Relational integrity + native vector search |
| **Storage** | MinIO (on-prem) / AWS S3 (cloud) | S3-compatible API |
| **Queue** | Redis | Background transcoding, AI analysis |

---

## 🚀 Quick Start (Docker Compose)

```bash
# 1. Clone and configure
git clone https://github.com/truemycornea/TestGHCP.git aura
cd aura
cp .env.example .env
# ↑ Edit .env and set JWT_SECRET, POSTGRES_PASSWORD, MINIO_ROOT_PASSWORD

# 2. Start all services (Postgres+pgvector, Redis, MinIO, Backend, Frontend)
docker compose up -d

# 3. Open the app
open http://localhost:3000
```

The database schema (including pgvector indices) is applied automatically on first boot via `docker-entrypoint-initdb.d`.

---

## 🛠 Local Development

### Backend

```bash
cd backend

# Install Go 1.22+
# https://go.dev/dl/

# Run with live reload (optional: install air)
go run ./cmd/api

# Build static binary
CGO_ENABLED=0 go build -o aura-backend ./cmd/api
```

### Frontend

```bash
cd frontend
npm install
npm run dev       # http://localhost:3000
npm run type-check
```

---

## 🗄 Database Schema

The schema (`backend/migrations/000001_init_schema.sql`) creates:

| Table | Purpose |
|-------|---------|
| `users` | Accounts with RBAC roles; supports local + OIDC auth |
| `assets` | Photos & videos with non-destructive storage references |
| `tags` / `asset_tags` | User and AI-generated keyword labels |
| `albums` / `album_assets` | Named collections |
| `shared_links` | Expirable, password-protected share URLs |
| `persons` / `faces` | Face clustering and recognition data |
| `job_log` | Background worker audit trail |

### pgvector Indices

```sql
-- Semantic search: cosine similarity on 512-dim CLIP embeddings
CREATE INDEX idx_assets_clip_embedding
    ON assets USING ivfflat (clip_embedding vector_cosine_ops)
    WITH (lists = 100);

-- Face clustering: HNSW for high-recall nearest-neighbour search
CREATE INDEX idx_faces_embedding
    ON faces USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
```

---

## 📡 API Reference

Interactive Swagger UI is served at `http://localhost:8080/swagger/index.html` when running the backend.

### Key Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/auth/register` | Register a new user |
| `POST` | `/api/v1/auth/login` | Obtain JWT token pair |
| `GET` | `/api/v1/users/me` | Current user profile |
| `POST` | `/api/v1/assets/upload` | Upload photo or video |
| `GET` | `/api/v1/assets` | List assets (filterable) |
| `GET` | `/api/v1/assets/:id` | Get asset + presigned URL |
| `PATCH`| `/api/v1/assets/:id/favourite` | Toggle favourite |
| `DELETE`|`/api/v1/assets/:id` | Trash asset |
| `GET` | `/api/v1/search?q=...` | Semantic natural-language search |

---

## 🔒 Security

- Passwords are hashed with **bcrypt** (cost 12).
- All API endpoints (except `/auth/*`) require a **Bearer JWT**.
- JWT secrets must be set via environment variable (`JWT_SECRET`).
- Object storage URLs are served via **time-limited pre-signed URLs** (1 hour default).
- Share links support **per-link bcrypt password hashing** and **expiry timestamps**.

---

## 📄 License

MIT
