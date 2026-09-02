package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

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
func setupDB() {
	os.Remove("./" + os.Getenv("TEST_DB_NAME") + ".db")
	initDB(true)
}
