package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shik3i/koalanotes/internal/db"
	"github.com/shik3i/koalanotes/internal/middleware"
)

var (
	testJWTSecret = []byte("test-secret-for-testing-only")
)

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()
	dir, err := os.MkdirTemp("", "koalanotes-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func decodeJSON(t *testing.T, body *bytes.Buffer, v any) {
	t.Helper()
	if err := json.NewDecoder(body).Decode(v); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
}

func registerAccount(t *testing.T, database *db.DB, email, password string) *httptest.ResponseRecorder {
	t.Helper()
	body := RegisterRequest{Email: email, Password: password}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		t.Fatalf("failed to encode register body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler := NewRegisterHandler(database, testJWTSecret)
	handler(rec, req)
	return rec
}

func loginAccount(t *testing.T, database *db.DB, email, password string) *httptest.ResponseRecorder {
	t.Helper()
	body := LoginRequest{Email: email, Password: password}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		t.Fatalf("failed to encode login body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler := NewLoginHandler(database, testJWTSecret)
	handler(rec, req)
	return rec
}

func pushBlobs(t *testing.T, database *db.DB, token string, blobs []db.BlobRecord) *httptest.ResponseRecorder {
	t.Helper()
	body := PushRequest{Blobs: blobs}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		t.Fatalf("failed to encode push body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/sync/push", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	authMw := middleware.Auth(testJWTSecret)
	authMw(http.HandlerFunc(NewSyncPushHandler(database))).ServeHTTP(rec, req)
	return rec
}

func pullBlobs(t *testing.T, database *db.DB, token, since string) *httptest.ResponseRecorder {
	t.Helper()
	path := "/api/sync/pull"
	if since != "" {
		path += "?since=" + since
	}
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	authMw := middleware.Auth(testJWTSecret)
	authMw(http.HandlerFunc(NewSyncPullHandler(database))).ServeHTTP(rec, req)
	return rec
}

func decodeAuthResponse(t *testing.T, rec *httptest.ResponseRecorder) AuthResponse {
	t.Helper()
	var resp AuthResponse
	decodeJSON(t, rec.Body, &resp)
	return resp
}

func decodePushResponse(t *testing.T, rec *httptest.ResponseRecorder) PushResponse {
	t.Helper()
	var resp PushResponse
	decodeJSON(t, rec.Body, &resp)
	return resp
}

func decodePullResponse(t *testing.T, rec *httptest.ResponseRecorder) PullResponse {
	t.Helper()
	var resp PullResponse
	decodeJSON(t, rec.Body, &resp)
	return resp
}

func decodeErrorResponse(t *testing.T, rec *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()
	var resp ErrorResponse
	decodeJSON(t, rec.Body, &resp)
	return resp
}

// ---- Auth Tests ----

func TestRegister_Success(t *testing.T) {
	database := setupTestDB(t)
	rec := registerAccount(t, database, "test@example.com", "secure-password")

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	resp := decodeAuthResponse(t, rec)
	if resp.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", resp.Email)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.AccountID == "" {
		t.Error("expected non-empty account_id")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "dup@example.com", "password123")

	rec := registerAccount(t, database, "dup@example.com", "password123")
	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409 for duplicate email, got %d", rec.Code)
	}

	errResp := decodeErrorResponse(t, rec)
	if errResp.Error == "" {
		t.Error("expected error message for duplicate email")
	}
}

func TestRegister_EmptyFields(t *testing.T) {
	database := setupTestDB(t)
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"empty email", "", "password"},
		{"empty password", "test@example.com", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := registerAccount(t, database, tt.email, tt.password)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "login@example.com", "the-password")

	rec := loginAccount(t, database, "login@example.com", "the-password")
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := decodeAuthResponse(t, rec)
	if resp.Email != "login@example.com" {
		t.Errorf("expected email 'login@example.com', got '%s'", resp.Email)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "wrongpw@example.com", "correct-password")

	rec := loginAccount(t, database, "wrongpw@example.com", "wrong-password")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for wrong password, got %d", rec.Code)
	}
}

func TestLogin_NonexistentEmail(t *testing.T) {
	database := setupTestDB(t)
	rec := loginAccount(t, database, "nobody@example.com", "some-password")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for nonexistent email, got %d", rec.Code)
	}
}

func TestLogin_EmailCaseInsensitive(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "CaseTest@Example.com", "password")

	// Login with different case
	rec := loginAccount(t, database, "casetest@example.com", "password")
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for case-insensitive login, got %d", rec.Code)
	}
}

// ---- Sync Tests ----

var testBlobCounter int64

func newTestBlob(campaignID, payload string) db.BlobRecord {
	testBlobCounter++
	return db.BlobRecord{
		ID:               fmt.Sprintf("blob-%s-%d", campaignID, testBlobCounter),
		CampaignKeyID:    campaignID,
		EncryptedPayload: payload,
		VectorClock:      "2026-01-01T00:00:00Z",
	}
}

func TestPush_Unauthenticated(t *testing.T) {
	database := setupTestDB(t)
	req := httptest.NewRequest(http.MethodPost, "/api/sync/push", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler := NewSyncPushHandler(database)
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestPushPull_RoundTrip(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "sync@example.com", "password")
	loginRec := loginAccount(t, database, "sync@example.com", "password")

	loginResp := decodeAuthResponse(t, loginRec)
	token := loginResp.Token

	// Push two blobs
	blobs := []db.BlobRecord{
		newTestBlob("campaign-1", `{"iv":"aGVsbG8=","ciphertext":"d29ybGQ="}`),
		newTestBlob("campaign-2", `{"iv":"d29ybGQ=","ciphertext":"aGVsbG8="}`),
	}

	pushRec := pushBlobs(t, database, token, blobs)
	if pushRec.Code != http.StatusOK {
		t.Errorf("expected status 200 for push, got %d", pushRec.Code)
	}

	pushResp := decodePushResponse(t, pushRec)
	if pushResp.Accepted != 2 {
		t.Errorf("expected 2 accepted blobs, got %d", pushResp.Accepted)
	}

	// Pull all blobs
	pullRec := pullBlobs(t, database, token, "")
	if pullRec.Code != http.StatusOK {
		t.Errorf("expected status 200 for pull, got %d", pullRec.Code)
	}

	pullResp := decodePullResponse(t, pullRec)
	if len(pullResp.Blobs) != 2 {
		t.Errorf("expected 2 blobs in pull, got %d", len(pullResp.Blobs))
	}
}

func TestPushPull_Incremental(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "incr@example.com", "password")
	loginRec := loginAccount(t, database, "incr@example.com", "password")

	loginResp := decodeAuthResponse(t, loginRec)
	token := loginResp.Token

	// Push initial blob
	blobs := []db.BlobRecord{
		newTestBlob("campaign-incr", `{"iv":"aGVsbG8=","ciphertext":"d29ybGQ="}`),
	}
	pushBlobs(t, database, token, blobs)

	// Pull with a since time before creation
	pullRec := pullBlobs(t, database, token, "2025-01-01T00:00:00Z")
	if pullRec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", pullRec.Code)
	}
	pullResp := decodePullResponse(t, pullRec)
	if len(pullResp.Blobs) != 1 {
		t.Errorf("expected 1 blob with old since, got %d", len(pullResp.Blobs))
	}

	// Pull with a since time after creation (should get nothing)
	pullRec2 := pullBlobs(t, database, token, "2099-01-01T00:00:00Z")
	pullResp2 := decodePullResponse(t, pullRec2)
	if len(pullResp2.Blobs) != 0 {
		t.Errorf("expected 0 blobs with future since, got %d", len(pullResp2.Blobs))
	}
}

func TestPushPull_IsolatedByAccount(t *testing.T) {
	database := setupTestDB(t)

	// Register two accounts
	registerAccount(t, database, "alice@example.com", "password")
	aliceLogin := loginAccount(t, database, "alice@example.com", "password")
	alice := decodeAuthResponse(t, aliceLogin)

	registerAccount(t, database, "bob@example.com", "password")
	bobLogin := loginAccount(t, database, "bob@example.com", "password")
	bob := decodeAuthResponse(t, bobLogin)

	// Alice pushes a blob
	pushBlobs(t, database, alice.Token, []db.BlobRecord{
		newTestBlob("alice-campaign", `{"iv":"aA==","ciphertext":"Yg=="}`),
	})

	// Bob pulls — should get nothing
	bobPull := pullBlobs(t, database, bob.Token, "")
	bobPullResp := decodePullResponse(t, bobPull)
	if len(bobPullResp.Blobs) != 0 {
		t.Errorf("expected bob to see 0 blobs, got %d", len(bobPullResp.Blobs))
	}

	// Alice pulls — should get her blob
	alicePull := pullBlobs(t, database, alice.Token, "")
	alicePullResp := decodePullResponse(t, alicePull)
	if len(alicePullResp.Blobs) != 1 {
		t.Errorf("expected alice to see 1 blob, got %d", len(alicePullResp.Blobs))
	}
}

func TestPush_UpdatesExistingBlob(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "update@example.com", "password")
	loginRec := loginAccount(t, database, "update@example.com", "password")
	loginResp := decodeAuthResponse(t, loginRec)
	token := loginResp.Token

	// Push initial
	pushBlobs(t, database, token, []db.BlobRecord{
		{
			ID:               "blob-campaign-upd",
			CampaignKeyID:    "campaign-upd",
			EncryptedPayload: `{"iv":"b2xk","ciphertext":"ZGF0YQ=="}`,
			VectorClock:      "2026-01-01T00:00:00Z",
		},
	})

	// Push update with different payload (same ID as initial push)
	pushBlobs(t, database, token, []db.BlobRecord{
		{
			ID:               "blob-campaign-upd",
			CampaignKeyID:    "campaign-upd",
			EncryptedPayload: `{"iv":"bmV3","ciphertext":"ZGF0YQ=="}`,
			VectorClock:      "2026-06-01T00:00:00Z",
		},
	})

	// Pull — should have 1 blob with the new payload
	pullRec := pullBlobs(t, database, token, "")
	pullResp := decodePullResponse(t, pullRec)
	if len(pullResp.Blobs) != 1 {
		t.Fatalf("expected 1 blob after update, got %d", len(pullResp.Blobs))
	}
	if pullResp.Blobs[0].EncryptedPayload != `{"iv":"bmV3","ciphertext":"ZGF0YQ=="}` {
		t.Errorf("expected updated payload, got '%s'", pullResp.Blobs[0].EncryptedPayload)
	}
}

func TestPull_Unauthenticated(t *testing.T) {
	database := setupTestDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/sync/pull", nil)
	rec := httptest.NewRecorder()
	handler := NewSyncPullHandler(database)
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// ---- Error-path Tests ----

func TestRegister_InvalidEmail(t *testing.T) {
	database := setupTestDB(t)
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"missing @", "invalid", "password123"},
		{"too long email", strings.Repeat("a", 255) + "@b.com", "password123"},
		{"short password", "test@example.com", "short"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := registerAccount(t, database, tt.email, tt.password)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
			errResp := decodeErrorResponse(t, rec)
			if errResp.Error == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}

func TestRegister_MalformedJSON(t *testing.T) {
	database := setupTestDB(t)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		strings.NewReader(`{not json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	NewRegisterHandler(database, testJWTSecret)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestLogin_MalformedJSON(t *testing.T) {
	database := setupTestDB(t)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{not json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	NewLoginHandler(database, testJWTSecret)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPush_MalformedJSON(t *testing.T) {
	database := setupTestDB(t)
	// Need a valid token so middleware passes and handler reads the body
	registerAccount(t, database, "malformed@example.com", "password123")
	loginRec := loginAccount(t, database, "malformed@example.com", "password123")
	var loginResp AuthResponse
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/sync/push",
		strings.NewReader(`{not json`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	rec := httptest.NewRecorder()
	authMw := middleware.Auth(testJWTSecret)
	authMw(http.HandlerFunc(NewSyncPushHandler(database))).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for malformed json, got %d", rec.Code)
	}
}

func TestPush_InvalidToken(t *testing.T) {
	database := setupTestDB(t)
	req := httptest.NewRequest(http.MethodPost, "/api/sync/push",
		strings.NewReader(`{"blobs":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	authMw := middleware.Auth(testJWTSecret)
	authMw(http.HandlerFunc(NewSyncPushHandler(database))).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for invalid token, got %d", rec.Code)
	}
}

func TestPush_ExpiredToken(t *testing.T) {
	database := setupTestDB(t)
	// Generate a token that expired 1 hour ago
	claims := jwt.MapClaims{
		"sub":   "test-account",
		"email": "test@example.com",
		"iat":   time.Now().Add(-2 * time.Hour).Unix(),
		"exp":   time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(testJWTSecret)
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/sync/push",
		strings.NewReader(`{"blobs":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	authMw := middleware.Auth(testJWTSecret)
	authMw(http.HandlerFunc(NewSyncPushHandler(database))).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for expired token, got %d", rec.Code)
	}
}

func TestPull_InvalidSince(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "since@example.com", "password123")
	loginRec := loginAccount(t, database, "since@example.com", "password123")
	var loginResp AuthResponse
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login: %v", err)
	}

	// Pull with malformed since parameter
	rec := pullBlobs(t, database, loginResp.Token, "not-a-date")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid since, got %d", rec.Code)
	}

	// Pull with valid since should work
	rec2 := pullBlobs(t, database, loginResp.Token, "2025-01-01T00:00:00Z")
	if rec2.Code != http.StatusOK {
		t.Errorf("expected status 200 for valid since, got %d", rec2.Code)
	}
}

func TestPush_EmptyBlobs(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "empty@example.com", "password123")
	loginRec := loginAccount(t, database, "empty@example.com", "password123")
	var loginResp AuthResponse
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login: %v", err)
	}

	// Push with empty blobs array
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(PushRequest{Blobs: []db.BlobRecord{}})
	req := httptest.NewRequest(http.MethodPost, "/api/sync/push", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	rec := httptest.NewRecorder()
	authMw := middleware.Auth(testJWTSecret)
	authMw(http.HandlerFunc(NewSyncPushHandler(database))).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty blobs, got %d", rec.Code)
	}
}

func TestPush_MissingBlobFields(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "fields@example.com", "password123")
	loginRec := loginAccount(t, database, "fields@example.com", "password123")
	var loginResp AuthResponse
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login: %v", err)
	}

	tests := []struct {
		name  string
		blobs []db.BlobRecord
	}{
		{"missing id", []db.BlobRecord{{CampaignKeyID: "k", EncryptedPayload: "{}"}}},
		{"missing campaign_key_id", []db.BlobRecord{{ID: "b", EncryptedPayload: "{}"}}},
		{"missing payload", []db.BlobRecord{{ID: "b", CampaignKeyID: "k"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(PushRequest{Blobs: tt.blobs})
			req := httptest.NewRequest(http.MethodPost, "/api/sync/push", &buf)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+loginResp.Token)
			rec := httptest.NewRecorder()
			authMw := middleware.Auth(testJWTSecret)
			authMw(http.HandlerFunc(NewSyncPushHandler(database))).ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
		})
	}
}
