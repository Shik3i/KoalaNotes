package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	addr := os.Getenv("KOALA_LISTEN_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	dataDir := os.Getenv("KOALA_DATA_DIR")
	if dataDir == "" {
		dataDir = "/data"
	}

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

	// Future API routes placeholder
	// mux.HandleFunc("POST /api/auth/register", ...)
	// mux.HandleFunc("POST /api/auth/login", ...)
	// mux.HandleFunc("GET /api/sync/pull", ...)
	// mux.HandleFunc("POST /api/sync/push", ...)

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
