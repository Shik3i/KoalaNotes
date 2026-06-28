package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shik3i/koalanotes/internal/db"
	"github.com/shik3i/koalanotes/internal/middleware"
)

type PushRequest struct {
	Blobs []db.BlobRecord `json:"blobs"`
}

type PushResponse struct {
	Accepted int `json:"accepted"`
}

type PullResponse struct {
	Blobs []db.BlobRecord `json:"blobs"`
}

// NewSyncPushHandler returns a handler for POST /api/sync/push.
// Accepts an array of encrypted blob records and upserts them.
func NewSyncPushHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := middleware.GetAccountID(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
			return
		}

		var req PushRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}

		if len(req.Blobs) == 0 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "no blobs provided"})
			return
		}

		// Assign account_id to each blob
		for i := range req.Blobs {
			req.Blobs[i].AccountID = accountID
		}

		if err := database.UpsertBlobs(req.Blobs); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to store blobs"})
			return
		}

		writeJSON(w, http.StatusOK, PushResponse{Accepted: len(req.Blobs)})
	}
}

// NewSyncPullHandler returns a handler for GET /api/sync/pull.
// Returns all blob records for the authenticated account, optionally filtered by ?since=ISO8601.
func NewSyncPullHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := middleware.GetAccountID(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
			return
		}

		since := r.URL.Query().Get("since")

		blobs, err := database.GetBlobs(accountID, since)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch blobs"})
			return
		}

		if blobs == nil {
			blobs = []db.BlobRecord{}
		}

		writeJSON(w, http.StatusOK, PullResponse{Blobs: blobs})
	}
}
