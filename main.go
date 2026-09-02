package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var port int64
var config Config

func main() {

	loadEnv()
	loadConfig()

	log.Println("Starting server...")
	initDB(false)
	defer db.Close()

	mux := http.NewServeMux()
	mux.Handle("POST /api/scores/{gameID}", rateLimit(http.HandlerFunc(submitScore)))
	mux.HandleFunc("GET /api/leaderboard/{gameID}", leaderboardHandler)
	mux.HandleFunc("GET /api/leaderboard/{gameID}/{userID}", getUserScore)
	mux.HandleFunc("POST /api/games", addGame)
	mux.HandleFunc("GET /api/games", getGames)
	addr := fmt.Sprintf(":%d", port)
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
	var err error
	port, err = strconv.ParseInt(os.Getenv("PORT"), 10, 0)
	if err != nil {
		log.Fatal("No port set in .env", err)
		return
	}
}
func loadConfig() {
	// load config
	var config Config

}
