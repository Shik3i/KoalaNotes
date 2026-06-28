package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/shik3i/koalanotes/internal/db"
	"github.com/shik3i/koalanotes/internal/handler"
	"github.com/shik3i/koalanotes/internal/middleware"
)

const (
	defaultAddr = ":8080"
	serverName  = "KoalaNotes"
	version     = "0.0.1"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	addr := envOrDefault("KOALA_LISTEN_ADDR", defaultAddr)
	dataDir := envOrDefault("KOALA_DATA_DIR", "/data")
	jwtSecretStr := os.Getenv("KOALA_JWT_SECRET")
	if jwtSecretStr == "" {
		slog.Error("KOALA_JWT_SECRET environment variable is required")
		os.Exit(1)
	}
	if jwtSecretStr == "change-me-in-production" {
		slog.Error("KOALA_JWT_SECRET must not be the default value 'change-me-in-production'")
		os.Exit(1)
	}
	jwtSecret := []byte(jwtSecretStr)

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		slog.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	// Open database
	database, err := db.Open(filepath.Join(dataDir, "koalanotes.db"))
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	slog.Info("starting server",
		"name", serverName,
		"version", version,
		"addr", addr,
		"data_dir", dataDir,
	)

	mux := http.NewServeMux()

	// Health and version endpoints
	mux.HandleFunc("GET /healthz", handler.Healthz)
	mux.HandleFunc("GET /api/version", handler.Version(version))

	// Auth endpoints (no JWT required)
	mux.HandleFunc("POST /api/auth/register", handler.NewRegisterHandler(database, jwtSecret))
	mux.HandleFunc("POST /api/auth/login", handler.NewLoginHandler(database, jwtSecret))

	// Sync endpoints (JWT required)
	authMw := middleware.Auth(jwtSecret)
	mux.Handle("POST /api/sync/push", authMw(http.HandlerFunc(handler.NewSyncPushHandler(database))))
	mux.Handle("GET /api/sync/pull", authMw(http.HandlerFunc(handler.NewSyncPullHandler(database))))

	// Wrap with middleware
	wrapped := middleware.Logging(mux)
	wrapped = middleware.CORS(wrapped)

	srv := &http.Server{
		Addr:         addr,
		Handler:      wrapped,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		slog.Info("received signal, shutting down", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}()

	slog.Info(fmt.Sprintf("%s server listening at %s", serverName, addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
