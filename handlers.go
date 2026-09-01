package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func getScores(w http.ResponseWriter, r *http.Request) {
	//TODO: Implement
	http.Error(w, "Unimplemented", http.StatusNotImplemented)
}

func submitScore(w http.ResponseWriter, r *http.Request) {
	var req SubmitScoreRequest
	// decode body to req type
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	gameID := int64(1)
	// some input validation
	switch {
	case req.PlayerName == "":
		http.Error(w, "Missing player name", http.StatusBadRequest)
		return
	case req.Score < 0:
		http.Error(w, "Negative score", http.StatusBadRequest)
		return
	}

	// valid
	result, err := db.Exec(
		"INSERT INTO scores (game_id, player_name, score) VALUES (?, ?, ?)",
		gameID, req.PlayerName, req.Score,
	)
	if err != nil {
		http.Error(w, "Failed to save score", http.StatusInternalServerError)
	}

	scoreID, _ := result.LastInsertId()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      scoreID,
		"message": "Score submitted",
	})
}

func leaderboardHandler(w http.ResponseWriter, r *http.Request) {

	gameID := r.PathValue("gameID")

	if gameID == "" {
		http.Error(w, "missing game ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
	SELECT player_name, score, created_at,
		RANK() OVER (ORDER BY score DESC) as rank
	FROM scores
	WHERE game_id = ?
	ORDER BY score DESC
	LIMIT 10`,
		gameID,
	)

	if err != nil {
		http.Error(w, "failed to fetch leaderboard", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.PlayerName, &entry.Score, &entry.CreatedAt, &entry.Rank); err != nil {
			http.Error(w, "failed to parse scores", http.StatusInternalServerError)
			return
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "failed to read leaderboard rows", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(entries)
}
func addGame(w http.ResponseWriter, r *http.Request) {
	var game Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	game.CreatedAt = time.Now()
	data, _ := json.Marshal(game)
	http.Error(w, string(data), http.StatusNotImplemented)

}
