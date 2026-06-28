package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORS middleware adds configurable CORS headers.
// Reads allowed origins from KOALA_CORS_ORIGINS env var (comma-separated).
// Defaults to http://localhost:5173 if not set.
func CORS(next http.Handler) http.Handler {
	originsEnv := os.Getenv("KOALA_CORS_ORIGINS")
	allowed := strings.Split(originsEnv, ",")
	if len(allowed) == 1 && allowed[0] == "" {
		allowed = []string{"http://localhost:5173"}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowOrigin := "*"
		for _, a := range allowed {
			if a == origin || a == "*" {
				allowOrigin = origin
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
