package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"uuid"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	_ = godotenv.Load()
	if os.Getenv("TEST_DB_NAME") == "" {
		os.Setenv("TEST_DB_NAME", "leaderboard_test")
	}
	adminKey = "test-admin-secret"
	LDBConfig = Config{
		MaxScore:      10000000,
		MaxRequests:   30,
		Port:          8000,
		MaxNameLength: 24,
		MaxBodySize:   1048576,
	}
	os.Exit(m.Run())
}

func setupDB() {
	if db != nil {
		db.Close()
	}
	name := os.Getenv("TEST_DB_NAME")
	os.Remove("./" + name + ".db")
	os.Remove("./" + name + ".db-wal")
	os.Remove("./" + name + ".db-shm")
	initDB(true)
}

func TestCreateGame(t *testing.T) {
	setupDB()
	body := `{"name":"New Game"}`
	req := httptest.NewRequest(http.MethodPost, "/api/games", strings.NewReader(body))
	req.Header.Set("content-type", "application/json")

	rec := httptest.NewRecorder()
	addGame(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("should respond with 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var game Game
	json.Unmarshal(rec.Body.Bytes(), &game)

	switch {
	case game.Name != "New Game":
		t.Errorf("incorrect game name: %q", game.Name)
	case game.ID == 0:
		t.Errorf("expected an id, got zero")
	case game.APIKey == "":
		t.Errorf("missing API Key")
	}
}

// seedGame creates a game through the handler and returns it.
func seedGame(t *testing.T) Game {
	t.Helper()
	body := `{"name":"Test Game"}`
	req := httptest.NewRequest(http.MethodPost, "/api/games", strings.NewReader(body))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()
	addGame(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("seedGame failed: %d %s", rec.Code, rec.Body.String())
	}
	var game Game
	if err := json.Unmarshal(rec.Body.Bytes(), &game); err != nil {
		t.Fatalf("failed to decode seeded game: %v", err)
	}
	return game
}

func TestGetGames(t *testing.T) {
	setupDB()
	seedGame(t)

	req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
	rec := httptest.NewRecorder()
	getGames(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var games []Game
	if err := json.Unmarshal(rec.Body.Bytes(), &games); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(games))
	}
	if games[0].Name != "Test Game" {
		t.Errorf("unexpected name: %q", games[0].Name)
	}
	if games[0].ID == 0 {
		t.Errorf("expected non-zero id")
	}
	if games[0].APIKey == "" {
		t.Errorf("expected api_key to be populated")
	}
}

func TestGetGamesEmpty(t *testing.T) {
	setupDB()

	req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
	rec := httptest.NewRecorder()
	getGames(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("expected empty list [], got %q", got)
	}
}

// scoreHash computes the client-side integrity hash:
// hex sha256 of "player_name|player_id|score|api_key".
// This mirrors what a game client must compute for the "hash" field.
func scoreHash(name, playerID string, score int64, apiKey string) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%s|%s|%d|%s", name, playerID, score, apiKey))
	return hex.EncodeToString(sum[:])
}

// submitScoreReq submits a score through the handler and returns the recorder.
func submitScoreReq(t *testing.T, game Game, playerID, name string, score int64) *httptest.ResponseRecorder {
	t.Helper()
	hash := scoreHash(name, playerID, score, game.APIKey)
	body := fmt.Sprintf(`{"player_id":%q,"player_name":%q,"score":%d,"hash":%q}`, playerID, name, score, hash)
	req := httptest.NewRequest(http.MethodPost, "/api/scores/"+strconv.FormatInt(game.ID, 10), strings.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("X-API-Key", game.APIKey)
	req.SetPathValue("gameID", strconv.FormatInt(game.ID, 10))
	rec := httptest.NewRecorder()
	submitScore(rec, req)
	return rec
}

func TestSubmitScore(t *testing.T) {
	setupDB()
	game := seedGame(t)

	rec := submitScoreReq(t, game, uuid.New().String(), "Alice", 100)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp SubmitScoreResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !resp.Accepted {
		t.Errorf("expected accepted, got %+v", resp)
	}
	if resp.SubmittedScore != 100 {
		t.Errorf("expected submitted 100, got %d", resp.SubmittedScore)
	}
	if resp.CurrentBest != 0 {
		t.Errorf("expected previous best 0, got %d", resp.CurrentBest)
	}
}

func TestSubmitScoreBadKey(t *testing.T) {
	setupDB()
	game := seedGame(t)

	body := `{"player_id":"` + uuid.New().String() + `","player_name":"Alice","score":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/scores/"+strconv.FormatInt(game.ID, 10), strings.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("X-API-Key", "wrong-key")
	req.SetPathValue("gameID", strconv.FormatInt(game.ID, 10))

	rec := httptest.NewRecorder()
	submitScore(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSubmitScoreMissingKey(t *testing.T) {
	setupDB()
	game := seedGame(t)

	body := `{"player_id":"` + uuid.New().String() + `","player_name":"Alice","score":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/scores/"+strconv.FormatInt(game.ID, 10), strings.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.SetPathValue("gameID", strconv.FormatInt(game.ID, 10))

	rec := httptest.NewRecorder()
	submitScore(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSubmitScoreBadHash(t *testing.T) {
	setupDB()
	game := seedGame(t)

	body := fmt.Sprintf(`{"player_id":%q,"player_name":%q,"score":%d,"hash":%q}`,
		uuid.New().String(), "Alice", 100, strings.Repeat("0", 64))
	req := httptest.NewRequest(http.MethodPost, "/api/scores/"+strconv.FormatInt(game.ID, 10), strings.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("X-API-Key", game.APIKey)
	req.SetPathValue("gameID", strconv.FormatInt(game.ID, 10))

	rec := httptest.NewRecorder()
	submitScore(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLeaderboard(t *testing.T) {
	setupDB()
	game := seedGame(t)

	submitScoreReq(t, game, uuid.New().String(), "Alice", 100)
	submitScoreReq(t, game, uuid.New().String(), "Bob", 200)

	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard/"+strconv.FormatInt(game.ID, 10), nil)
	req.SetPathValue("gameID", strconv.FormatInt(game.ID, 10))
	rec := httptest.NewRecorder()
	leaderboardHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Total-Count"); got != "2" {
		t.Errorf("expected X-Total-Count 2, got %q", got)
	}
	var entries []LeaderboardEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].PlayerName != "Bob" || entries[0].Score != 200 {
		t.Errorf("expected Bob(200) first, got %+v", entries[0])
	}
	if entries[1].PlayerName != "Alice" || entries[1].Score != 100 {
		t.Errorf("expected Alice(100) second, got %+v", entries[1])
	}
	if entries[0].Rank != 1 || entries[1].Rank != 2 {
		t.Errorf("unexpected ranks: %d, %d", entries[0].Rank, entries[1].Rank)
	}
}

func TestLeaderboardPagination(t *testing.T) {
	setupDB()
	game := seedGame(t)

	submitScoreReq(t, game, uuid.New().String(), "Alice", 100)
	submitScoreReq(t, game, uuid.New().String(), "Bob", 200)
	submitScoreReq(t, game, uuid.New().String(), "Carol", 300)

	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard/"+strconv.FormatInt(game.ID, 10)+"?limit=2&offset=1", nil)
	req.SetPathValue("gameID", strconv.FormatInt(game.ID, 10))
	rec := httptest.NewRecorder()
	leaderboardHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Total-Count"); got != "3" {
		t.Errorf("expected X-Total-Count 3, got %q", got)
	}
	var entries []LeaderboardEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].PlayerName != "Bob" || entries[1].PlayerName != "Alice" {
		t.Errorf("unexpected page: %+v, %+v", entries[0], entries[1])
	}
}

func TestGetUserScore(t *testing.T) {
	setupDB()
	game := seedGame(t)
	playerID := uuid.New().String()
	submitScoreReq(t, game, playerID, "Alice", 150)

	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard/"+strconv.FormatInt(game.ID, 10)+"/"+playerID, nil)
	req.SetPathValue("gameID", strconv.FormatInt(game.ID, 10))
	req.SetPathValue("userID", playerID)
	rec := httptest.NewRecorder()
	getUserScore(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var entry LeaderboardEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entry); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if entry.PlayerName != "Alice" || entry.Score != 150 {
		t.Errorf("unexpected entry: %+v", entry)
	}
	if entry.Rank != 1 {
		t.Errorf("expected rank 1, got %d", entry.Rank)
	}
}

func TestGetUserScoreNotFound(t *testing.T) {
	setupDB()
	game := seedGame(t)

	userID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard/"+strconv.FormatInt(game.ID, 10)+"/"+userID, nil)
	req.SetPathValue("gameID", strconv.FormatInt(game.ID, 10))
	req.SetPathValue("userID", userID)
	rec := httptest.NewRecorder()
	getUserScore(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRequireAdmin(t *testing.T) {
	handler := requireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	cases := []struct {
		name   string
		key    string
		status int
	}{
		{"missing key", "", http.StatusUnauthorized},
		{"wrong key", "nope", http.StatusUnauthorized},
		{"correct key", adminKey, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.key != "" {
				req.Header.Set("X-Admin-Key", tc.key)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Errorf("expected %d, got %d", tc.status, rec.Code)
			}
		})
	}
}
