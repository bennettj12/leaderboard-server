package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func corsHeaders(rec *httptest.ResponseRecorder) map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  rec.Header().Get("Access-Control-Allow-Origin"),
		"Access-Control-Allow-Methods": rec.Header().Get("Access-Control-Allow-Methods"),
		"Access-Control-Allow-Headers": rec.Header().Get("Access-Control-Allow-Headers"),
	}
}

func TestCORSPreflight(t *testing.T) {
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/scores/1", nil)
	req.Header.Set("Origin", "https://html-classic.itch.zone")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type, x-api-key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d", rec.Code)
	}
	for k, v := range corsHeaders(rec) {
		if v == "" {
			t.Errorf("preflight missing %q header", k)
		}
	}
}

func TestCORSOnNormalRequest(t *testing.T) {
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard/1", nil)
	req.Header.Set("Origin", "https://html-classic.itch.zone")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", got)
	}
}
