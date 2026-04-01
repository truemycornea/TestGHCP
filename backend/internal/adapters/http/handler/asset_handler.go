package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	asset_service "github.com/truemycornea/aura/backend/internal/application/asset"
	"github.com/truemycornea/aura/backend/internal/domain/asset"
	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
)

// AssetHandler handles asset upload, retrieval and management endpoints.
type AssetHandler struct {
	svc    *asset_service.Service
	logger *zap.Logger
}

// NewAssetHandler creates a new AssetHandler.
func NewAssetHandler(svc *asset_service.Service, logger *zap.Logger) *AssetHandler {
	return &AssetHandler{svc: svc, logger: logger}
}

// RegisterRoutes attaches endpoints to the authenticated router group.
func (h *AssetHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/assets/upload", h.Upload)
	r.GET("/assets", h.List)
	r.GET("/assets/:id", h.Get)
	r.PATCH("/assets/:id/favourite", h.Favourite)
	r.DELETE("/assets/:id", h.Trash)
}

// Upload godoc
//
//	@Summary		Upload a new photo or video
//	@Tags			assets
//	@Security		BearerAuth
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"Media file"
//	@Success		201	{object}	asset.Asset
//	@Failure		400	{object}	apierrors.APIError
//	@Router			/assets/upload [post]
func (h *AssetHandler) Upload(c *gin.Context) {
	ownerID := mustUserID(c)

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, apierrors.New(http.StatusBadRequest, "file field required", err.Error()))
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, apierrors.New(http.StatusInternalServerError, apierrors.ErrInternalServer.Error(), ""))
		return
	}
	defer f.Close()

	a, err := h.svc.UploadAsset(c.Request.Context(), ownerID, fh.Filename, f, fh.Size)
	if err != nil {
		status := apierrors.HTTPStatusFor(err)
		c.JSON(status, apierrors.New(status, err.Error(), ""))
		return
	}

	c.JSON(http.StatusCreated, a)
}

// List godoc
//
//	@Summary		List assets with optional filters
//	@Tags			assets
//	@Security		BearerAuth
//	@Produce		json
//	@Param			limit		query	int		false	"Page size (default 50)"
//	@Param			offset		query	int		false	"Page offset"
//	@Param			media_type	query	string	false	"photo|video"
//	@Success		200	{object}	listAssetsResponse
//	@Router			/assets [get]
func (h *AssetHandler) List(c *gin.Context) {
	ownerID := mustUserID(c)
	limit := queryInt(c, "limit", 50)
	offset := queryInt(c, "offset", 0)

	filter := asset.ListFilter{}
	if mt := c.Query("media_type"); mt != "" {
		m := asset.MediaType(mt)
		filter.MediaType = &m
	}

	assets, total, err := h.svc.ListAssets(c.Request.Context(), ownerID, filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apierrors.New(http.StatusInternalServerError, apierrors.ErrInternalServer.Error(), ""))
		return
	}

	c.JSON(http.StatusOK, listAssetsResponse{Assets: assets, Total: total, Limit: limit, Offset: offset})
}

// Get godoc
//
//	@Summary		Get a single asset with its presigned download URL
//	@Tags			assets
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path	string	true	"Asset UUID"
//	@Success		200	{object}	getAssetResponse
//	@Failure		404	{object}	apierrors.APIError
//	@Router			/assets/{id} [get]
func (h *AssetHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apierrors.New(http.StatusBadRequest, "invalid asset id", ""))
		return
	}

	a, url, err := h.svc.GetAsset(c.Request.Context(), id)
	if err != nil {
		status := apierrors.HTTPStatusFor(err)
		c.JSON(status, apierrors.New(status, err.Error(), ""))
		return
	}

	c.JSON(http.StatusOK, getAssetResponse{Asset: a, DownloadURL: url})
}

// Favourite godoc
//
//	@Summary		Toggle favourite flag on an asset
//	@Tags			assets
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string				true	"Asset UUID"
//	@Param			body	body	favouriteRequest	true	"Favourite flag"
//	@Success		204
//	@Router			/assets/{id}/favourite [patch]
func (h *AssetHandler) Favourite(c *gin.Context) {
	ownerID := mustUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apierrors.New(http.StatusBadRequest, "invalid asset id", ""))
		return
	}

	var req favouriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.New(http.StatusBadRequest, apierrors.ErrBadRequest.Error(), err.Error()))
		return
	}

	if err := h.svc.FavouriteAsset(c.Request.Context(), id, ownerID, req.Favourite); err != nil {
		status := apierrors.HTTPStatusFor(err)
		c.JSON(status, apierrors.New(status, err.Error(), ""))
		return
	}

	c.Status(http.StatusNoContent)
}

// Trash godoc
//
//	@Summary		Move an asset to the trash
//	@Tags			assets
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Asset UUID"
//	@Success		204
//	@Router			/assets/{id} [delete]
func (h *AssetHandler) Trash(c *gin.Context) {
	ownerID := mustUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, apierrors.New(http.StatusBadRequest, "invalid asset id", ""))
		return
	}

	if err := h.svc.TrashAsset(c.Request.Context(), id, ownerID); err != nil {
		status := apierrors.HTTPStatusFor(err)
		c.JSON(status, apierrors.New(status, err.Error(), ""))
		return
	}

	c.Status(http.StatusNoContent)
}

// ---- DTOs ----

type listAssetsResponse struct {
	Assets []*asset.Asset `json:"assets"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type getAssetResponse struct {
	*asset.Asset
	DownloadURL string `json:"download_url"`
}

type favouriteRequest struct {
	Favourite bool `json:"favourite"`
}

// ---- helpers ----

func mustUserID(c *gin.Context) uuid.UUID {
	raw, _ := c.Get("user_id")
	id, _ := uuid.Parse(raw.(string))
	return id
}

func queryInt(c *gin.Context, key string, def int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return def
	}
	return v
}
