package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

type requestMinute struct {
	count uint
	start time.Time
}
type reqPartial struct {
	PlayerID string `json:"player_id"`
}

var users map[string]requestMinute = make(map[string]requestMinute)
var mu sync.Mutex

/*
Expects request to include a player_id and rate limits based on that value
checks for rate limits and then executes the handler
*/
func rateLimit(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get and restore body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed read", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		var req reqPartial
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "request missing ID", http.StatusBadRequest)
			return
		}
		mu.Lock()
		if val, ok := users[req.PlayerID]; ok {
			if diff := time.Since(val.start); diff >= time.Minute {
				val = requestMinute{0, time.Now()}
			}
			// exists
			val.count += 1
			if val.count > LDBConfig.MaxRequests {
				http.Error(w, "rate limited", http.StatusTooManyRequests)
				mu.Unlock()
				return
			}
			users[req.PlayerID] = val
		} else {
			// does not exist
			users[req.PlayerID] = requestMinute{1, time.Now()}
		}
		mu.Unlock()
		n.ServeHTTP(w, r)

	})
}

func startRateLimitCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			for id, w := range users {
				if time.Since(w.start) >= time.Minute {
					delete(users, id)
				}
			}
			mu.Unlock()
		}

	}()
}
