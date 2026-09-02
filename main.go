package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var LDBConfig Config

func main() {

	loadEnv()
	LDBConfig = loadConfig()

	log.Println("Starting server...")
	initDB(false)
	defer db.Close()

	mux := http.NewServeMux()
	mux.Handle("POST /api/scores/{gameID}", rateLimit(http.HandlerFunc(submitScore)))
	mux.HandleFunc("GET /api/leaderboard/{gameID}", leaderboardHandler)
	mux.HandleFunc("GET /api/leaderboard/{gameID}/{userID}", getUserScore)
	mux.HandleFunc("POST /api/games", addGame)
	mux.HandleFunc("GET /api/games", getGames)
	addr := fmt.Sprintf(":%d", LDBConfig.Port)
	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))

}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed loading .env", err)
		return
	} else {
		log.Println(".ENV Loaded:", os.Getenv("ENVLOADED"))
	}
}
func loadConfig() Config {
	var config Config
	// load config
	data, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("Missing config.json", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("Error parsing config.json", err)
	}
	return config
}
