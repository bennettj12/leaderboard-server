package main

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB(testDatabase bool) {
	var err error

	dbName := os.Getenv("DB_NAME")
	if testDatabase {
		dbName = os.Getenv("TEST_DB_NAME")
	}

	db, err = sql.Open("sqlite", "./"+dbName+".db")
	if err != nil {
		log.Fatal("Failed to open db.", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("failed to ping db", err)
	}

	createTables()
}

func createTables() {
	gamesTable := `
    CREATE TABLE IF NOT EXISTS games (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        api_key TEXT NOT NULL UNIQUE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`
	scoresTable := `
    CREATE TABLE IF NOT EXISTS scores (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        game_id INTEGER NOT NULL,
        player_name TEXT NOT NULL,
				player_id TEXT NOT NULL,
        score INTEGER NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (game_id) REFERENCES games(id),
        UNIQUE(game_id, player_id)
    );`
	scoreIndex := `
    CREATE INDEX IF NOT EXISTS idx_scores_game_score 
    ON scores(game_id, score DESC);`

	statements := []string{gamesTable, scoresTable, scoreIndex}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			log.Fatal("Failed to execute statement:", err)
		}
	}
}
