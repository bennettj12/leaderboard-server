package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"uuid"
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

	gameID := r.PathValue("gameID")
	// make sure it exists
	var game Game
	row := db.QueryRow(`SELECT * FROM games WHERE id = ?`, gameID)

	if err := row.Scan(&game.ID, &game.Name, &game.APIKey, &game.CreatedAt); err != nil {
		log.Print(err)
		http.Error(w, "failed to parse game information", http.StatusInternalServerError)
		return
	}
	if game.APIKey != req.APIKey {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
		return
	}
	// some input validation
	switch {
	case req.PlayerName == "":
		http.Error(w, "Missing player name", http.StatusBadRequest)
		return
	case len(req.PlayerName) > 24:
		http.Error(w, "Name too long (max 24 characters)", http.StatusBadRequest)
	case req.PlayerID == "":
		http.Error(w, "Missing player name", http.StatusBadRequest)
		return
	case req.Score < 0:
		http.Error(w, "Negative score", http.StatusBadRequest)
		return
	}

	// valid
	result, err := db.Exec(
		`INSERT INTO scores (game_id, player_name, player_id, score) VALUES (?, ?, ?, ?)
		ON CONFLICT(game_id, player_id) 
		DO UPDATE SET
			score = excluded.score,
			player_name = excluded.player_name,
			updated_at = CURRENT_TIMESTAMP
		WHERE excluded.score > scores.score
		`,
		gameID, req.PlayerName, req.PlayerID, req.Score,
	)
	if err != nil {
		log.Print(err)
		http.Error(w, "Failed to save score", http.StatusInternalServerError)
		return
	}

	scoreID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	message := "Score submitted"
	if rowsAffected == 0 {
		message = "Higher score already exists"
	}

	json.NewEncoder(w).Encode(map[string]any{
		"id":      scoreID,
		"message": message,
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
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
func addGame(w http.ResponseWriter, r *http.Request) {
	var game Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if game.Name == "" {
		http.Error(w, "Name must not be empty", http.StatusBadRequest)
		return
	}
	game.APIKey = uuid.New().String()
	game.CreatedAt = time.Now()
	result, err := db.Exec(`INSERT INTO games (name, created_at, api_key) VALUES (?, ?, ?)`, game.Name, game.CreatedAt, game.APIKey)
	if err != nil {
		http.Error(w, "Failed to write data", http.StatusInternalServerError)
		return
	}
	game.ID, err = result.LastInsertId()
	if err != nil {
		http.Error(w, "error getting new ID", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(game)

}
func getGames(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
