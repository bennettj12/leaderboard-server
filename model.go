package main

import "time"

// Data model
type Game struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}
type Score struct {
	ID         int64     `json:"id"`
	GameID     int64     `json:"game_id"`
	PlayerName string    `json:"player_name"`
	PlayerID   string    `json:"player_id"`
	Score      int64     `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// request/response model

type SubmitScoreRequest struct {
	PlayerName string `json:"player_name"`
	PlayerID   string `json:"player_id"`
	Score      int64  `json:"score"`
	APIKey     string `json:"api_key"`
}

type LeaderboardEntry struct {
	PlayerName string    `json:"player_name"`
	Score      int64     `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
	Rank       int       `json:"rank"`
}
