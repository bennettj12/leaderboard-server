package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	mux := http.NewServeMux()
	mux.Handle("POST /api/scores/{gameID}", rateLimit(http.HandlerFunc(submitScore)))
	mux.HandleFunc("GET /api/leaderboard/{gameID}", leaderboardHandler)
	mux.HandleFunc("GET /api/leaderboard/{gameID}/{userID}", getUserScore)
	mux.Handle("POST /api/games", requireAdmin(http.HandlerFunc(addGame)))
	mux.Handle("GET /api/games", requireAdmin(http.HandlerFunc(getGames)))
	addr := fmt.Sprintf(":%d", LDBConfig.Port)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go startRateLimitCleanup(ctx, time.Minute) // clear expired rate limits
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server starting on %s", addr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		log.Fatalf("server error: %v", err)
	case <-ctx.Done():
		log.Println("shutdown signal received")
	}
	// select passed : shutting down
	// wait a bit
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown incomplete: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("error closing database: %v", err)
	}
	log.Println("Shutdown complete.")

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
