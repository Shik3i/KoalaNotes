package handler

import (
	"encoding/json"
	"net/http"
)

// VersionResponse is the JSON response for the version endpoint.
type VersionResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Version returns an HTTP handler that responds with the server version.
func Version(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(VersionResponse{
			Name:    "KoalaNotes",
			Version: version,
		})
	}
}
