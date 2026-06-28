package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/shik3i/koalanotes/internal/db"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// fakeHash is a pre-computed bcrypt hash used to prevent timing-based email enumeration.
// Generated once at init with the same cost as real registration hashes.
var fakeHash string
var fakeHashOnce sync.Once

func getFakeHash() string {
	fakeHashOnce.Do(func() {
		h, err := bcrypt.GenerateFromPassword([]byte("timing-fake"), bcrypt.DefaultCost+2)
		if err != nil {
			panic("failed to generate fake bcrypt hash: " + err.Error())
		}
		fakeHash = string(h)
	})
	return fakeHash
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token     string `json:"token"`
	AccountID string `json:"account_id"`
	Email     string `json:"email"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// NewRegisterHandler returns a handler for POST /api/auth/register.
func NewRegisterHandler(database *db.DB, jwtSecret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if err := validateCredentials(req.Email, req.Password); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost+2)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "server error"})
			return
		}

		id, err := newUUID()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "server error"})
			return
		}

		// Atomic INSERT — no TOCTOU race, UNIQUE constraint catches duplicates
		if err := database.CreateAccount(r.Context(), id, req.Email, string(hash)); err != nil {
			if isUniqueConstraintErr(err) {
				writeJSON(w, http.StatusConflict, ErrorResponse{Error: "email already registered"})
			} else {
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create account"})
			}
			return
		}

		token, err := generateJWT(id, req.Email, jwtSecret)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
			return
		}

		writeJSON(w, http.StatusCreated, AuthResponse{Token: token, AccountID: id, Email: req.Email})
	}
}

// NewLoginHandler returns a handler for POST /api/auth/login.
func NewLoginHandler(database *db.DB, jwtSecret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if err := validateCredentials(req.Email, req.Password); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		id, passwordHash, err := database.GetAccountByEmail(r.Context(), req.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Perform fake bcrypt to prevent email enumeration via timing
				bcrypt.CompareHashAndPassword([]byte(getFakeHash()), []byte(req.Password))
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid email or password"})
			} else {
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
			}
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid email or password"})
			return
		}

		token, err := generateJWT(id, req.Email, jwtSecret)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
			return
		}

		writeJSON(w, http.StatusOK, AuthResponse{Token: token, AccountID: id, Email: req.Email})
	}
}

func validateCredentials(email, password string) error {
	if email == "" || password == "" {
		return errors.New("email and password are required")
	}
	if len(email) > 254 {
		return errors.New("email is too long")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 128 {
		return errors.New("password is too long")
	}
	if len(password) > 72 {
		return errors.New("password exceeds maximum length for bcrypt (72 bytes)")
	}
	return nil
}

func generateJWT(accountID, email string, secret []byte) (string, error) {
	claims := jwt.MapClaims{
		"sub":   accountID,
		"email": email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("json encode failed", "error", err, "status", status)
	}
}

// newUUID generates a v4 UUID string using crypto/rand.
func newUUID() (string, error) {
	var u [16]byte
	if _, err := rand.Read(u[:]); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	// Set version 4 bits
	u[6] = (u[6] & 0x0f) | 0x40
	// Set variant bits (10xx)
	u[8] = (u[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16]), nil
}

// isUniqueConstraintErr checks for SQLite UNIQUE constraint violation.
func isUniqueConstraintErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
