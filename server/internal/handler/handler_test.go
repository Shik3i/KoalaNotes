package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

// ---- Auth Tests ----

func TestRegister_Success(t *testing.T) {
	database := setupTestDB(t)
	rec := registerAccount(t, database, "test@example.com", "secure-password")

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

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

	var errResp ErrorResponse
	json.NewDecoder(rec.Body).Decode(&errResp)
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

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

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

func newTestBlob(campaignID, payload string) db.BlobRecord {
	return db.BlobRecord{
		ID:               campaignID,
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

	var loginResp AuthResponse
	json.NewDecoder(loginRec.Body).Decode(&loginResp)
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

	var pushResp PushResponse
	json.NewDecoder(pushRec.Body).Decode(&pushResp)
	if pushResp.Accepted != 2 {
		t.Errorf("expected 2 accepted blobs, got %d", pushResp.Accepted)
	}

	// Pull all blobs
	pullRec := pullBlobs(t, database, token, "")
	if pullRec.Code != http.StatusOK {
		t.Errorf("expected status 200 for pull, got %d", pullRec.Code)
	}

	var pullResp PullResponse
	json.NewDecoder(pullRec.Body).Decode(&pullResp)
	if len(pullResp.Blobs) != 2 {
		t.Errorf("expected 2 blobs in pull, got %d", len(pullResp.Blobs))
	}
}

func TestPushPull_Incremental(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "incr@example.com", "password")
	loginRec := loginAccount(t, database, "incr@example.com", "password")

	var loginResp AuthResponse
	json.NewDecoder(loginRec.Body).Decode(&loginResp)
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
	var pullResp PullResponse
	json.NewDecoder(pullRec.Body).Decode(&pullResp)
	if len(pullResp.Blobs) != 1 {
		t.Errorf("expected 1 blob with old since, got %d", len(pullResp.Blobs))
	}

	// Pull with a since time after creation (should get nothing)
	pullRec2 := pullBlobs(t, database, token, "2099-01-01T00:00:00Z")
	json.NewDecoder(pullRec2.Body).Decode(&pullResp)
	if len(pullResp.Blobs) != 0 {
		t.Errorf("expected 0 blobs with future since, got %d", len(pullResp.Blobs))
	}
}

func TestPushPull_IsolatedByAccount(t *testing.T) {
	database := setupTestDB(t)

	// Register two accounts
	registerAccount(t, database, "alice@example.com", "password")
	aliceLogin := loginAccount(t, database, "alice@example.com", "password")
	var alice AuthResponse
	json.NewDecoder(aliceLogin.Body).Decode(&alice)

	registerAccount(t, database, "bob@example.com", "password")
	bobLogin := loginAccount(t, database, "bob@example.com", "password")
	var bob AuthResponse
	json.NewDecoder(bobLogin.Body).Decode(&bob)

	// Alice pushes a blob
	pushBlobs(t, database, alice.Token, []db.BlobRecord{
		newTestBlob("alice-campaign", `{"iv":"aA==","ciphertext":"Yg=="}`),
	})

	// Bob pulls — should get nothing
	bobPull := pullBlobs(t, database, bob.Token, "")
	var bobPullResp PullResponse
	json.NewDecoder(bobPull.Body).Decode(&bobPullResp)
	if len(bobPullResp.Blobs) != 0 {
		t.Errorf("expected bob to see 0 blobs, got %d", len(bobPullResp.Blobs))
	}

	// Alice pulls — should get her blob
	alicePull := pullBlobs(t, database, alice.Token, "")
	var alicePullResp PullResponse
	json.NewDecoder(alicePull.Body).Decode(&alicePullResp)
	if len(alicePullResp.Blobs) != 1 {
		t.Errorf("expected alice to see 1 blob, got %d", len(alicePullResp.Blobs))
	}
}

func TestPush_UpdatesExistingBlob(t *testing.T) {
	database := setupTestDB(t)
	registerAccount(t, database, "update@example.com", "password")
	loginRec := loginAccount(t, database, "update@example.com", "password")
	var loginResp AuthResponse
	json.NewDecoder(loginRec.Body).Decode(&loginResp)
	token := loginResp.Token

	// Push initial
	pushBlobs(t, database, token, []db.BlobRecord{
		newTestBlob("campaign-upd", `{"iv":"b2xk","ciphertext":"ZGF0YQ=="}`),
	})

	// Push update with different payload
	pushBlobs(t, database, token, []db.BlobRecord{
		{
			ID:               "campaign-upd",
			CampaignKeyID:    "campaign-upd",
			EncryptedPayload: `{"iv":"bmV3","ciphertext":"ZGF0YQ=="}`,
			VectorClock:      "2026-06-01T00:00:00Z",
		},
	})

	// Pull — should have 1 blob with the new payload
	pullRec := pullBlobs(t, database, token, "")
	var pullResp PullResponse
	json.NewDecoder(pullRec.Body).Decode(&pullResp)
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
