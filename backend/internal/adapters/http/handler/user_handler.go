// Package handler contains the Gin HTTP handlers for the Aura API.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	user_service "github.com/truemycornea/aura/backend/internal/application/user"
	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
)

// UserHandler handles user-facing HTTP endpoints (auth + profile).
type UserHandler struct {
	svc    *user_service.Service
	logger *zap.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc *user_service.Service, logger *zap.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

// RegisterRoutes attaches endpoints to the provided router group.
func (h *UserHandler) RegisterRoutes(public, protected gin.IRouter) {
	public.POST("/auth/register", h.Register)
	public.POST("/auth/login", h.Login)
	protected.GET("/users/me", h.Me)
}

// Register godoc
//
//	@Summary		Register a new user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		registerRequest	true	"Registration payload"
//	@Success		201		{object}	userResponse
//	@Failure		400		{object}	apierrors.APIError
//	@Failure		409		{object}	apierrors.APIError
//	@Router			/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.New(http.StatusBadRequest, apierrors.ErrBadRequest.Error(), err.Error()))
		return
	}

	u, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		status := apierrors.HTTPStatusFor(err)
		c.JSON(status, apierrors.New(status, err.Error(), ""))
		return
	}

	c.JSON(http.StatusCreated, userResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        string(u.Role),
	})
}

// Login godoc
//
//	@Summary		Authenticate and obtain JWT tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest	true	"Login credentials"
//	@Success		200		{object}	user_service.TokenPair
//	@Failure		401		{object}	apierrors.APIError
//	@Router			/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.New(http.StatusBadRequest, apierrors.ErrBadRequest.Error(), err.Error()))
		return
	}

	tokens, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, apierrors.New(http.StatusUnauthorized, apierrors.ErrUnauthorized.Error(), ""))
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Me godoc
//
//	@Summary		Get the currently authenticated user
//	@Tags			users
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	userResponse
//	@Failure		401	{object}	apierrors.APIError
//	@Router			/users/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	rawID, _ := c.Get("user_id")
	userID, _ := rawID.(string)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, apierrors.New(http.StatusUnauthorized, apierrors.ErrUnauthorized.Error(), ""))
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, apierrors.New(http.StatusUnauthorized, apierrors.ErrUnauthorized.Error(), ""))
		return
	}

	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		status := apierrors.HTTPStatusFor(err)
		c.JSON(status, apierrors.New(status, err.Error(), ""))
		return
	}

	c.JSON(http.StatusOK, userResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        string(u.Role),
	})
}

// ---- request / response DTOs ----

type registerRequest struct {
	Email       string `json:"email"        binding:"required,email"`
	Password    string `json:"password"     binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type userResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}
