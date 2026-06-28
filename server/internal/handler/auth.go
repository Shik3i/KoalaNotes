package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/shik3i/koalanotes/internal/db"
)

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
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" || req.Password == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "email and password are required"})
			return
		}

		exists, err := database.AccountExists(req.Email)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "server error"})
			return
		}
		if exists {
			writeJSON(w, http.StatusConflict, ErrorResponse{Error: "email already registered"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "server error"})
			return
		}

		id := newUUID()
		if err := database.CreateAccount(id, req.Email, string(hash)); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create account"})
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
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" || req.Password == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "email and password are required"})
			return
		}

		id, passwordHash, err := database.GetAccountByEmail(req.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
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
	json.NewEncoder(w).Encode(v)
}

// newUUID generates a v4 UUID string using crypto/rand.
func newUUID() string {
	var u [16]byte
	if _, err := rand.Read(u[:]); err != nil {
		panic(fmt.Sprintf("failed to read random bytes: %v", err))
	}
	// Set version 4 bits
	u[6] = (u[6] & 0x0f) | 0x40
	// Set variant bits (10xx)
	u[8] = (u[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}
