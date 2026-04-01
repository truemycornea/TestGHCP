// Package search implements the semantic / natural-language search use-case.
package search

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/truemycornea/aura/backend/internal/domain/asset"
)

// EmbeddingPort generates a CLIP embedding vector for a text query.
// Concrete implementations may call a local Ollama model or the OpenAI API.
type EmbeddingPort interface {
	TextEmbedding(ctx context.Context, text string) ([]float32, error)
}

// Service implements the semantic search use-case.
type Service struct {
	assetRepo asset.Repository
	embedder  EmbeddingPort
	logger    *zap.Logger
}

// New creates a new search Service.
func New(assetRepo asset.Repository, embedder EmbeddingPort, logger *zap.Logger) *Service {
	return &Service{assetRepo: assetRepo, embedder: embedder, logger: logger}
}

// SemanticSearch converts a natural-language query into a CLIP embedding and
// retrieves the top-k most similar assets from pgvector.
func (s *Service) SemanticSearch(ctx context.Context, query string, limit int) ([]*asset.Asset, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	vec, err := s.embedder.TextEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("generating embedding for query %q: %w", query, err)
	}

	s.logger.Debug("semantic search", zap.String("query", query), zap.Int("vec_dim", len(vec)))

	results, err := s.assetRepo.SemanticSearch(ctx, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}

	return results, nil
}
