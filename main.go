package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/spamntaters/boot.dev-chirpy/internal/api"
	"github.com/spamntaters/boot.dev-chirpy/internal/database"
	"github.com/spamntaters/boot.dev-chirpy/internal/handlers"

	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	const filePathRoot = "."
	const port = "8080"

	dbURL := mustGetenv("DB_URL")
	platform := getEnvOrDefault("PLATFORM", "production")
	secret := mustGetenv("SECRET")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	cfg := &api.Config{
		DB:       database.New(db),
		Platform: platform,
		Secret:   secret,
	}

	mux := setupRoutes(cfg, filePathRoot)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filePathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func setupRoutes(cfg *api.Config, filePathRoot string) *http.ServeMux {
	mux := http.NewServeMux()

	// File server with metrics middleware
	fileServer := http.FileServer(http.Dir(filePathRoot))
	mux.Handle("GET /app/", cfg.MiddlewareMetrics(http.StripPrefix("/app", fileServer)))

	// Admin routes
	mux.HandleFunc("GET /admin/metrics", handlers.HandleMetrics(cfg))
	mux.HandleFunc("POST /admin/reset", handlers.HandleResetUsers(cfg))

	// Health check
	mux.HandleFunc("GET /api/healthz", handlers.HandleHealth)

	// User routes
	mux.HandleFunc("POST /api/users", handlers.HandleCreateUser(cfg))
	mux.HandleFunc("PUT /api/users", handlers.HandleUpdateUser(cfg))
	mux.HandleFunc("POST /api/login", handlers.HandleLogin(cfg))
	mux.HandleFunc("POST /api/refresh", handlers.HandleRefreshToken(cfg))
	mux.HandleFunc("POST /api/revoke", handlers.HandleRevokeToken(cfg))

	// Chirp routes
	mux.HandleFunc("POST /api/chirps", handlers.HandleCreateChirp(cfg))
	mux.HandleFunc("GET /api/chirps", handlers.HandleGetAllChirps(cfg))
	mux.HandleFunc("GET /api/chirps/{chirpID}", handlers.HandleGetChirpByID(cfg))
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", handlers.HandleDeleteChirpByID(cfg))

	// Polka webook
	mux.HandleFunc("POST /api/polka/webhooks", handlers.HandlePolkaEvent(cfg))

	return mux
}

func mustGetenv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
