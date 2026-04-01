// Aura API Server — entry point.
//
// @title			Aura API
// @version			1.0
// @description		High-performance, self-hostable image & video management platform.
// @termsOfService	http://swagger.io/terms/
// @contact.name	Aura Support
// @license.name	MIT
// @host			localhost:8080
// @BasePath		/api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	album_service "github.com/truemycornea/aura/backend/internal/application/album"
	asset_service "github.com/truemycornea/aura/backend/internal/application/asset"
	"github.com/truemycornea/aura/backend/internal/application/search"
	user_service "github.com/truemycornea/aura/backend/internal/application/user"
	"github.com/truemycornea/aura/backend/internal/adapters/http/handler"
	"github.com/truemycornea/aura/backend/internal/adapters/http/middleware"
	"github.com/truemycornea/aura/backend/internal/adapters/queue"
	pgrepo "github.com/truemycornea/aura/backend/internal/adapters/repository/postgres"
	"github.com/truemycornea/aura/backend/internal/adapters/storage"
	"github.com/truemycornea/aura/backend/pkg/config"
	"github.com/truemycornea/aura/backend/pkg/logger"
)

func main() {
	// ── Configuration ──────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// ── Logger ─────────────────────────────────────────────────────────────────
	log := logger.Must(cfg.Server.Environment)
	defer log.Sync() //nolint:errcheck

	// ── Database ────────────────────────────────────────────────────────────────
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatal("failed to connect to PostgreSQL", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("PostgreSQL ping failed", zap.Error(err))
	}
	log.Info("connected to PostgreSQL")

	// ── Redis ───────────────────────────────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis ping failed", zap.Error(err))
	}
	log.Info("connected to Redis")

	// ── Object Storage ──────────────────────────────────────────────────────────
	s3Storage, err := storage.NewS3Storage(ctx, cfg.Storage)
	if err != nil {
		log.Fatal("failed to init S3 storage", zap.Error(err))
	}

	// ── Adapters ────────────────────────────────────────────────────────────────
	userRepo := pgrepo.NewUserRepository(pool)
	assetRepo := pgrepo.NewAssetRepository(pool)
	faceRepo := pgrepo.NewFaceRepository(pool)
	_ = faceRepo // used by face service (omitted for brevity in this file)

	taskQueue := queue.NewRedisQueue(redisClient)

	// ── Application Services ────────────────────────────────────────────────────
	userSvc := user_service.New(userRepo, cfg.Auth, log)
	assetSvc := asset_service.New(assetRepo, s3Storage, taskQueue, log)
	albumRepo := pgrepo.NewAlbumRepository(pool)
	albumSvc := album_service.New(albumRepo, log)
	_ = albumSvc

	// Semantic search uses a stub embedder until Ollama/OpenAI is wired up.
	searchSvc := search.New(assetRepo, &stubEmbedder{}, log)

	// ── HTTP Handlers ────────────────────────────────────────────────────────────
	userHandler := handler.NewUserHandler(userSvc, log)
	assetHandler := handler.NewAssetHandler(assetSvc, log)
	searchHandler := handler.NewSearchHandler(searchSvc, log)

	// ── Gin Router ───────────────────────────────────────────────────────────────
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger(log))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "version": "1.0.0"})
	})

	v1 := r.Group("/api/v1")

	// Public routes (no auth required).
	userHandler.RegisterRoutes(v1, v1.Group("", middleware.Auth(userSvc, log)))

	// Authenticated routes.
	authed := v1.Group("", middleware.Auth(userSvc, log))
	assetHandler.RegisterRoutes(authed)
	searchHandler.RegisterRoutes(authed)

	// ── HTTP Server ───────────────────────────────────────────────────────────────
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info("starting Aura API server", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server…")
	shutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("server forced shutdown", zap.Error(err))
	}
	log.Info("server stopped")
}

// stubEmbedder returns a zero vector — replace with Ollama/OpenAI adapter.
type stubEmbedder struct{}

func (s *stubEmbedder) TextEmbedding(_ context.Context, _ string) ([]float32, error) {
	return make([]float32, 512), nil
}
