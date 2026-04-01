package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/truemycornea/aura/backend/internal/application/search"
	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
)

// SearchHandler handles semantic natural-language search endpoints.
type SearchHandler struct {
	svc    *search.Service
	logger *zap.Logger
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(svc *search.Service, logger *zap.Logger) *SearchHandler {
	return &SearchHandler{svc: svc, logger: logger}
}

// RegisterRoutes attaches search endpoints to the authenticated router group.
func (h *SearchHandler) RegisterRoutes(r gin.IRouter) {
	r.GET("/search", h.Semantic)
}

// Semantic godoc
//
//	@Summary		Semantic natural-language asset search
//	@Tags			search
//	@Security		BearerAuth
//	@Produce		json
//	@Param			q		query	string	true	"Natural-language query, e.g. 'sunset over mountains'"
//	@Param			limit	query	int		false	"Max results (default 50, max 200)"
//	@Success		200	{object}	semanticSearchResponse
//	@Failure		400	{object}	apierrors.APIError
//	@Router			/search [get]
func (h *SearchHandler) Semantic(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, apierrors.New(http.StatusBadRequest, "query parameter 'q' is required", ""))
		return
	}

	limit := queryInt(c, "limit", 50)
	results, err := h.svc.SemanticSearch(c.Request.Context(), q, limit)
	if err != nil {
		h.logger.Error("semantic search failed", zap.String("query", q), zap.Error(err))
		c.JSON(http.StatusInternalServerError, apierrors.New(http.StatusInternalServerError, apierrors.ErrInternalServer.Error(), ""))
		return
	}

	c.JSON(http.StatusOK, semanticSearchResponse{Results: results, Count: len(results)})
}

type semanticSearchResponse struct {
	Results interface{} `json:"results"`
	Count   int         `json:"count"`
}
