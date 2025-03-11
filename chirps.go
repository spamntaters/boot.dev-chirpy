package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/spamntaters/boot.dev-chirpy/internal/auth"
	"github.com/spamntaters/boot.dev-chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleAddChirp(resWriter http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(resWriter, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(resWriter, http.StatusUnauthorized, "Failed to find bearer token", err)
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(resWriter, http.StatusUnauthorized, "Failed to validate token", err)
		return
	}
	params.UserID = userId
	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams(params))
	if err != nil {
		respondWithError(resWriter, http.StatusBadRequest, "Bad Request", err)
		return
	}

	respondWithJSON(resWriter, http.StatusCreated, Chirp(chirp))
}

func (cfg *apiConfig) handleGetAllChirps(resWriter http.ResponseWriter, req *http.Request) {
	data, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		respondWithError(resWriter, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	var chirps []Chirp
	for _, chirp := range data {
		chirps = append(chirps, Chirp(chirp))
	}
	respondWithJSON(resWriter, http.StatusOK, chirps)
}

func (cfg *apiConfig) handleGetChirpByID(resWriter http.ResponseWriter, req *http.Request) {

	param := req.PathValue("chirpID")
	id, err := uuid.Parse(param)
	if err != nil {
		respondWithError(resWriter, http.StatusBadRequest, "invalid uuid", err)
		return
	}
	chirp, err := cfg.db.GetChirpByID(req.Context(), id)
	if err != nil {
		respondWithError(resWriter, http.StatusNotFound, "chirp not found", err)
	}

	respondWithJSON(resWriter, http.StatusOK, Chirp(chirp))
}
