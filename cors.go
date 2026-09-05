package main

import (
	"net/http"
)

// CORS headers for browser clients (e.g. the HTML5/Godot web build on itch.io).
//
// itch.io serves uploaded games from its own CDN origin (html-classic.itch.zone)
// inside an iframe, so the Origin we see is dynamic and not worth hardcoding.
// A wildcard is safe here: we don't use cookies or other credentials, and the
// per-game API key is embedded in the public client anyway, so allowing any
// origin adds no real risk. Requests without custom headers still need the
// response headers set, and requests with a JSON body + X-API-Key trigger an
// OPTIONS preflight that must not fall through to the method-specific routes.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Short-circuit browser preflights before routing.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
