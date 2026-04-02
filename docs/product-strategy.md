# Aura Competitive Analysis & Delivery Plan

This document maps Aura's objectives against leading photo/video platforms (Google Photos, ACDSee, Immich), identifies gaps, and outlines a phased roadmap to deliver an enterprise-grade, aesthetically pleasing experience. Each phase highlights how GitHub Copilot/automation can accelerate delivery.

## Current Objective Snapshot
- Self-hostable, high-performance media library with AI search, face grouping, deduplication, hybrid on-prem/cloud AI, and partner sharing (per README feature grid).
- Stack: Next.js 14 + Tailwind/Radix UI frontend; Go (Gin) backend with PostgreSQL + pgvector, Redis, S3/MinIO.

## Benchmarks & Gaps
| Capability | Google Photos | ACDSee | Immich | Aura (today) | Gap / Opportunity |
|------------|---------------|--------|--------|---------------|-------------------|
| Core upload & sync | Mature multi-device sync, live albums | Desktop-first ingest, cataloging | Mobile + web sync; self-hosted | Boilerplate ingestion | Offline-friendly background uploads, resumable multi-part, bandwidth controls |
| Organization | Auto albums, Memories, map, timeline, stories | Deep tagging, ratings, IPTC/EXIF editing | Albums, map, people, sharing | Basic listing/filtering | Hierarchical albums, smart collections, saved searches, timeline/map views |
| Search & AI | Best-in-class semantic + visual | Keyword/metadata search | CLIP-based search, faces | Semantic search planned | Multi-modal search (text+image), relevance tuning, search analytics |
| Editing | Non-destructive edits, filters | Pro-grade edits, batch tools | Basic edits | Not implemented | Inline edits, presets, reversible adjustments with history |
| Sharing | Shared libraries, partner sharing, link controls | Limited cloud share; mostly export | Shared albums, links | Boilerplate partner sharing | Granular permissions, watermarking, expiration presets, viewer analytics |
| Privacy & governance | Consumer-grade; limited admin | Single-user DAM; no multi-tenant | Self-hosted; basic auth | JWT auth, RBAC boilerplate | Multi-tenant IAM, audit logs, DLP, retention/legal hold, SSO/SAML |
| Extensibility | Closed | Plugins, metadata tools | Open source extensions | None | Plugin/app layer (webhooks, workers), admin UI for integrations |

## Unique Selling Propositions to Stand Out
- **Hybrid AI control**: Local Ollama + cloud providers with per-tenant policies, cost guards, and explainable tagging.
- **Enterprise governance**: Multi-tenant RBAC, SSO/SAML/OIDC, audit trails, retention/legal hold, DLP rules, watermarking, and eDiscovery export.
- **Performance at scale**: Virtualized gallery (already present), tiered storage (hot/cold), adaptive thumbnails/proxies, GPU-aware pipelines.
- **Privacy-by-design**: On-prem processing, zero-data-sharing mode, private spaces, and per-link passwords/expiry by default.
- **Automation-first**: Policy-driven ingestion (auto-tag, auto-rotate, denoise), smart dedup with human-in-the-loop merge, and background health checks.

## Phased Delivery (Stage-by-Stage)
**Stage 0 – Foundations (infra + auth)**  
Scope: Harden JWT/OIDC/SAML, seed RBAC roles, storage config validation, health checks, observability (structured logs, tracing), rate limiting.  
Copilot/automation: Generate config validation, middleware scaffolds, Terraform snippets, and unit test skeletons.

**Stage 1 – Ingestion & Metadata Backbone**  
Scope: Chunked/resumable uploads, background checksum + pHash, EXIF/ IPTC extraction, sidecar writes, task queue retries, upload clients (web + mobile dropzone), metadata editing.  
Copilot/automation: Generate upload handlers, queue jobs, and typed API clients; propose edge-case tests (large files, network drops).

**Stage 2 – Intelligence Layer**  
Scope: CLIP embedding pipeline, pgvector tuning, face detection/clustering, smart dedup suggestions, geo reverse-geocoding, AI captions, NSFW detection with policy toggles.  
Copilot/automation: Draft model wrappers, SQL migrations, and evaluation notebooks; suggest prompts for explainable tags.

**Stage 3 – Discovery & Collaboration**  
Scope: Semantic + faceted search (text/image/date/location/device/people), saved searches, smart albums, map + timeline views, shared albums/partner libraries, comments/reactions, presence indicators.  
Copilot/automation: Generate search DTOs, pagination adapters, and UI state machines; author contract tests for search relevance.

**Stage 4 – Enterprise & Governance**  
Scope: Tenancy isolation, SSO/SAML with just-in-time provisioning, admin console, audit logging, retention/legal hold, DLP rules, watermarking, eDiscovery export, approval workflows, org-wide quotas.  
Copilot/automation: Scaffold policy engines, log redaction filters, and RBAC checks; generate compliance evidence scripts (access logs queries).

**Stage 5 – Experience Polish & Apps**  
Scope: Mobile-friendly gestures, offline-first caching, background sync progress, immersive lightbox with edits/history, batch operations, keyboard shortcuts, notification center, theming.  
Copilot/automation: Suggest responsive layout tweaks, generate accessibility checks, and automate Storybook-like fixture creation.

## UI & UX Improvement Plan
- **Onboarding & upload**: Guided setup for storage and AI providers; drag-drop with chunked progress, conflict resolution, and post-upload actions (auto-tag/auto-album).  
- **Gallery & browse**: Masonry/justified layout option; quick filters (favourites, people, places, device, file type), saved filter presets, infinite scroll + year/month jump list, map view toggle, and “Today/This week” highlight rows.  
- **Search & discovery**: Combined semantic + structured filters, typeahead with recent queries, “search by image” dropzone, and relevance explainers (why this result).  
- **Asset detail**: Edge-to-edge media, histogram/exif panel, edit history, related/dupe suggestions, faces/places/tags chips, and share insights (views, expiry).  
- **Sharing flows**: Explicit permissions (view/comment/edit/download), watermark + expiry presets, password and OTP options, branded share pages, and invite via emails/links.  
- **Admin & governance**: Tenant switcher, policy editor (retention/DLP), audit timeline, quota dashboards, integration connectors (webhooks, S3 buckets, Slack/Teams).  
- **Performance & accessibility**: Skeletons everywhere, virtualization already in place; add focus states, keyboard shortcuts, prefers-reduced-motion support, and lazy hydration for non-critical widgets.

## Implementation Notes
- Prioritize data models and APIs first (upload → metadata → AI → sharing), then layer UI polish to avoid rework.  
- Keep feature flags for AI providers and enterprise controls to enable gradual rollout.  
- Measure: ingest throughput, search latency, dedup precision/recall, face clustering accuracy, share link CTR, and crash-free sessions.
