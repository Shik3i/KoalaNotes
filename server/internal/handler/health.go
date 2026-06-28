package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/shik3i/koalanotes/internal/db"
)

// HealthResponse is the JSON response for the health check endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// NewHealthzHandler returns a handler that checks database connectivity.
func NewHealthzHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := database.Ping(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(HealthResponse{Status: "unhealthy"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}
}
