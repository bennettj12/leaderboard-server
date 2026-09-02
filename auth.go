package main

import (
	"crypto/subtle"
	"net/http"
)

// requireAdmin protects server-only endpoints (game creation/listing).
// The admin key is a server-side secret and must NOT be distributed to clients.
func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Admin-Key")
		if key == "" || subtle.ConstantTimeCompare([]byte(adminKey), []byte(key)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
