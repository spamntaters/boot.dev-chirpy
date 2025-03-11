package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/spamntaters/boot.dev-chirpy/internal/database"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}

func (cfg *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resWriter http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(resWriter, req)
	})
}

func (cfg *apiConfig) handlerMetrics(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	resWriter.WriteHeader(http.StatusOK)
	var res []byte
	resWriter.Write(fmt.Appendf(res, `
  <html>
    <body>
      <h1>Welcome, Chirpy Admin</h1>
      <p>Chirpy has been visited %d times!</p>
    </body>
  </html>
  `, cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handlerReset(resWriter http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	resWriter.WriteHeader(http.StatusOK)
}

func healthHandler(resWriter http.ResponseWriter, req *http.Request) {
	resWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resWriter.WriteHeader(http.StatusOK)
	resWriter.Write([]byte("OK"))
}

func validateChirpHandler(resWriter http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(req.Body)
	chirp := parameters{}
	err := decoder.Decode(&chirp)
	if err != nil {
		respondWithError(resWriter, http.StatusInternalServerError, "Something went wrong", err)
	}

	const maxChirpLength = 140
	if len(chirp.Body) > maxChirpLength {
		respondWithError(resWriter, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleanedBody := getCleanedBody(chirp.Body)
	respondWithJSON(resWriter, http.StatusOK, returnVals{CleanedBody: cleanedBody})
}

func getCleanedBody(body string) string {
	censorWords := []string{"kerfuffle", "sharbert", "fornax"}
	splitBody := strings.Split(body, " ")
	for i, word := range splitBody {
		if slices.Contains(censorWords, strings.ToLower(word)) {
			splitBody[i] = "****"
		}
	}
	return strings.Join(splitBody, " ")
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func main() {
	godotenv.Load()
	const filePathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	apiConfig := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:         secret,
	}

	mux.Handle("GET /app/", apiConfig.middlewareMetrics(http.StripPrefix("/app", (http.FileServer(http.Dir(filePathRoot))))))
	mux.HandleFunc("GET /admin/metrics", apiConfig.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiConfig.resetUsersHandler)
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	mux.HandleFunc("POST /api/users", apiConfig.addUserHandler)
	mux.HandleFunc("POST /api/login", apiConfig.handleUserLogin)
	mux.HandleFunc("POST /api/chirps", apiConfig.handleAddChirp)
	mux.HandleFunc("GET /api/chirps", apiConfig.handleGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.handleGetChirpByID)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filePathRoot, port)
	log.Fatal(server.ListenAndServe())
}
