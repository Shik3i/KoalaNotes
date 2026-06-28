package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shik3i/koalanotes/internal/db"
	"github.com/shik3i/koalanotes/internal/middleware"
)

const maxPushBody = 10 << 20 // 10 MB

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
func NewSyncPushHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := middleware.GetAccountID(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxPushBody)

		var req PushRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if IsMaxBytesError(err) {
				writeJSON(w, http.StatusRequestEntityTooLarge, ErrorResponse{Error: "request body too large"})
			} else {
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			}
			return
		}

		if len(req.Blobs) == 0 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "no blobs provided"})
			return
		}

		if len(req.Blobs) > 1000 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "too many blobs (max 1000)"})
			return
		}

		// Validate and assign account_id to each blob
		for i := range req.Blobs {
			b := &req.Blobs[i]
			if b.ID == "" {
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "blob id is required"})
				return
			}
			if b.CampaignKeyID == "" {
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "campaign_key_id is required"})
				return
			}
			if b.EncryptedPayload == "" {
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "encrypted_payload is required"})
				return
			}
			if len(b.EncryptedPayload) > maxPushBody {
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "encrypted_payload too large"})
				return
			}
			b.AccountID = accountID
		}

		if err := database.UpsertBlobs(r.Context(), req.Blobs); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to store blobs"})
			return
		}

		writeJSON(w, http.StatusOK, PushResponse{Accepted: len(req.Blobs)})
	}
}

// NewSyncPullHandler returns a handler for GET /api/sync/pull.
func NewSyncPullHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := middleware.GetAccountID(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
			return
		}

		since := r.URL.Query().Get("since")

		blobs, err := database.GetBlobs(r.Context(), accountID, since)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		if blobs == nil {
			blobs = []db.BlobRecord{}
		}

		writeJSON(w, http.StatusOK, PullResponse{Blobs: blobs})
	}
}

// IsMaxBytesError checks if an error is from http.MaxBytesReader.
func IsMaxBytesError(err error) bool {
	if err == nil {
		return false
	}
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}
