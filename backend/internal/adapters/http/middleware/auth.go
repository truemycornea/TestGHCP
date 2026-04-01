// Package middleware provides Gin middleware for the Aura API server.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/truemycornea/aura/backend/internal/domain/user"
	apierrors "github.com/truemycornea/aura/backend/pkg/errors"
)

// ContextKey constants for values stored in the Gin context.
const (
	ContextKeyUserID = "user_id"
	ContextKeyRole   = "user_role"
)

// TokenValidator abstracts JWT validation so middleware has no import cycle.
type TokenValidator interface {
	ValidateToken(token string) (userID, role string, err error)
}

// Auth extracts and validates the Bearer token from the Authorization header.
// On success it sets user_id and user_role in the Gin context.
func Auth(validator TokenValidator, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apierrors.New(
				http.StatusUnauthorized, apierrors.ErrUnauthorized.Error(), "missing or malformed Authorization header",
			))
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		userID, role, err := validator.ValidateToken(token)
		if err != nil {
			logger.Warn("invalid JWT", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, apierrors.New(
				http.StatusUnauthorized, apierrors.ErrUnauthorized.Error(), "",
			))
			return
		}

		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyRole, role)
		c.Next()
	}
}

// RequireRole aborts with 403 if the authenticated user does not have the required role.
func RequireRole(required user.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ContextKeyRole)
		if role != string(required) && role != string(user.RoleAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, apierrors.New(
				http.StatusForbidden, apierrors.ErrForbidden.Error(), "",
			))
			return
		}
		c.Next()
	}
}

// CORS sets permissive CORS headers for development; tighten for production.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// RequestLogger logs every request with method, path and latency.
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
		)
	}
}
