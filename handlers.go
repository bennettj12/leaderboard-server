package main

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
	"uuid"
)

// GET /api/leaderboard/{gameID}/{userID}
func getUserScore(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("gameID")
	userID := r.PathValue("userID")

	var ldb LeaderboardEntry

	rowScore := db.QueryRow(`
		SELECT player_name, score, created_at,
			(SELECT COUNT(*) + 1 FROM scores s2
			WHERE s2.game_id = s1.game_id AND s2.score > s1.score) AS rank
		FROM scores s1
		WHERE s1.game_id = ? AND s1.player_id = ?`, gameID, userID)

	if err := rowScore.Scan(&ldb.PlayerName, &ldb.Score, &ldb.CreatedAt, &ldb.Rank); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Failed to find score", http.StatusNotFound)
		} else {
			log.Println(err)
			http.Error(w, "Error scanning database", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ldb)
}

func submitScore(w http.ResponseWriter, r *http.Request) {
	var req SubmitScoreRequest
	// decode body to req type
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "Missing API key", http.StatusUnauthorized)
		return
	}

	gameID := r.PathValue("gameID")
	// make sure it exists
	var game Game
	row := db.QueryRow(`SELECT * FROM games WHERE id = ?`, gameID)

	if err := row.Scan(&game.ID, &game.Name, &game.APIKey, &game.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "game not found", http.StatusNotFound)
			return
		}
		log.Print(err)
		http.Error(w, "failed to parse game information", http.StatusInternalServerError)
		return
	}
	if subtle.ConstantTimeCompare([]byte(game.APIKey), []byte(apiKey)) != 1 {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}
	// some input validation
	if req.isValid(game.APIKey) == false {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var previousScore int64
	// get previous best
	row = db.QueryRow(
		`SELECT score FROM scores WHERE game_id = ? AND player_id = ?`, gameID, req.PlayerID)
	if err := row.Scan(&previousScore); err != nil {
		// No previous score found
		previousScore = 0
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

	var response SubmitScoreResponse
	response = SubmitScoreResponse{Accepted: true, SubmittedScore: req.Score, CurrentBest: previousScore}
	rowsAffected, _ := result.RowsAffected()

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if rowsAffected == 0 {
		response.Accepted = false
		response.Message = "Higher score already exists"
	} else {
		response.Message = "Score accepted"
	}

	json.NewEncoder(w).Encode(response)
}

func leaderboardHandler(w http.ResponseWriter, r *http.Request) {

	gameID := r.PathValue("gameID")

	if gameID == "" {
		http.Error(w, "missing game ID", http.StatusBadRequest)
		return
	}

	limit, offset := 10, 0

	q := r.URL.Query()
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = n
		if limit > 100 {
			limit = 100
		}
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = n
	}

	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM scores WHERE game_id = ?`, gameID).Scan(&total); err != nil {
		http.Error(w, "failed to count leaderboard", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(`
	SELECT player_name, score, created_at,
		RANK() OVER (ORDER BY score DESC) as rank
	FROM scores
	WHERE game_id = ?
	ORDER BY score DESC, player_id ASC
	LIMIT ? OFFSET ?`,
		gameID, limit, offset,
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
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
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
	rows, err := db.Query(`SELECT id, name, api_key, created_at FROM games ORDER BY id`)
	if err != nil {
		log.Print(err)
		http.Error(w, "failed to fetch games", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	games := []Game{}
	for rows.Next() {
		var game Game
		if err := rows.Scan(&game.ID, &game.Name, &game.APIKey, &game.CreatedAt); err != nil {
			log.Print(err)
			http.Error(w, "failed to parse games", http.StatusInternalServerError)
			return
		}
		games = append(games, game)
	}
	if err := rows.Err(); err != nil {
		log.Print(err)
		http.Error(w, "failed to read games", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(games)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	// 200 0K
}
