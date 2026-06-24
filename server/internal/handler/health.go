package handler

import (
	"encoding/json"
	"net/http"
)

// HealthResponse is the JSON response for the health check endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// Healthz responds with a simple health check.
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}
