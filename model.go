package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
	"uuid"
)

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
	Hash       string `json:"hash"`
}

func (s *SubmitScoreRequest) isValid(apiKey string) bool {
	switch {
	case s.Score > LDBConfig.MaxScore:
		return false
	case s.Score < 0:
		return false
	case s.PlayerName == "":
		return false
	case len(s.PlayerName) > int(LDBConfig.MaxNameLength):
		return false
	}
	if _, err := uuid.Parse(s.PlayerID); err != nil {
		return false
	}
	// compute hash
	// s.PlayerName + s.PlayerID + fmt.Sprint(s.Score) + apiKey
	hashBytes := sha256.Sum256(fmt.Appendf(nil, "%s|%s|%d|%s", s.PlayerName, s.PlayerID, s.Score, apiKey))
	hashString := hex.EncodeToString(hashBytes[:])
	if hashString != s.Hash {
		return false
	}

	return true
}

type SubmitScoreResponse struct {
	Accepted       bool   `json:"accepted"`
	Message        string `json:"message"`
	CurrentBest    int64  `json:"current_best"`
	SubmittedScore int64  `json:"submitted_score"`
}

type LeaderboardEntry struct {
	PlayerName string    `json:"player_name"`
	Score      int64     `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
	Rank       int       `json:"rank"`
}

// AdminScoreEntry is LeaderboardEntry plus player_id. player_id is deliberately
// kept out of the public leaderboard response; only admin endpoint uses this.
type AdminScoreEntry struct {
	PlayerName string    `json:"player_name"`
	PlayerID   string    `json:"player_id"`
	Score      int64     `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
}

// config model

type Config struct {
	MaxScore      int64 `json:"max_score"`
	MaxRequests   uint  `json:"max_requests"`
	Port          int64 `json:"port"`
	MaxNameLength uint  `json:"max_name_length"`
	MaxBodySize   int64 `json:"max_body_size"`
}
