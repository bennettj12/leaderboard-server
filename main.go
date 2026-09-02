package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed loading .env", err)
		return
	} else {
		log.Println(".ENV Loaded:", os.Getenv("ENVLOADED"))
	}
	var port, err = strconv.ParseInt(os.Getenv("PORT"), 10, 0)
	if err != nil {
		log.Fatal("No port set in .env", err)
		return
	}
	log.Println("Starting server...")
	initDB(false)
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/scores/{gameID}", submitScore)
	mux.HandleFunc("GET /api/leaderboard/{gameID}", leaderboardHandler)
	mux.HandleFunc("GET /api/leaderboard/{gameID}/{userID}", getUserScore)
	mux.HandleFunc("POST /api/games", addGame)
	mux.HandleFunc("GET /api/games", getGames)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))

}
