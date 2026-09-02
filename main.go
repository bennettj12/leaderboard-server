package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var LDBConfig Config
var adminKey string

func main() {

	loadEnv()
	LDBConfig = loadConfig()
	adminKey = os.Getenv("ADMIN_KEY")
	if adminKey == "" {
		log.Fatal("ADMIN_KEY not set in .env")
	}

	log.Println("Starting server...")
	initDB(false)
	defer db.Close()

	mux := http.NewServeMux()
	mux.Handle("POST /api/scores/{gameID}", rateLimit(http.HandlerFunc(submitScore)))
	mux.HandleFunc("GET /api/leaderboard/{gameID}", leaderboardHandler)
	mux.HandleFunc("GET /api/leaderboard/{gameID}/{userID}", getUserScore)
	mux.Handle("POST /api/games", requireAdmin(http.HandlerFunc(addGame)))
	mux.Handle("GET /api/games", requireAdmin(http.HandlerFunc(getGames)))
	addr := fmt.Sprintf(":%d", LDBConfig.Port)
	log.Printf("Server starting on %s", addr)

	go startRateLimitCleanup(time.Minute) // clear expired rate limits

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
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
