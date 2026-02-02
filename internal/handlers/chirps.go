package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/spamntaters/boot.dev-chirpy/internal/api"
	"github.com/spamntaters/boot.dev-chirpy/internal/auth"
	"github.com/spamntaters/boot.dev-chirpy/internal/database"
)

type ChirpResponse struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserID    string `json:"user_id"`
}

func HandleCreateChirp(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body string `json:"body"`
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			api.RespondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			api.RespondWithError(w, http.StatusUnauthorized, "Failed to find bearer token", err)
			return
		}
		userId, err := auth.ValidateJWT(token, cfg.Secret)
		if err != nil {
			api.RespondWithError(w, http.StatusUnauthorized, "Failed to validate token", err)
			return
		}
		chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   params.Body,
			UserID: userId,
		})
		if err != nil {
			api.RespondWithError(w, http.StatusBadRequest, "Bad Request", err)
			return
		}
		api.RespondWithJSON(w, http.StatusCreated, ChirpResponse{
			ID:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
			UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
			Body:      chirp.Body,
			UserID:    chirp.UserID.String(),
		})
	}
}

func HandleGetAllChirps(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := cfg.DB.GetAllChirps(r.Context())
		if err != nil {
			api.RespondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
		chirps := make([]ChirpResponse, len(data))
		for i, chirp := range data {
			chirps[i] = ChirpResponse{
				ID:        chirp.ID.String(),
				CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
				UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
				Body:      chirp.Body,
				UserID:    chirp.UserID.String(),
			}
		}
		api.RespondWithJSON(w, http.StatusOK, chirps)
	}
}

func HandleGetChirpByID(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		param := r.PathValue("chirpID")
		id, err := uuid.Parse(param)
		if err != nil {
			api.RespondWithError(w, http.StatusBadRequest, "Invalid uuid", err)
			return
		}
		chirp, err := cfg.DB.GetChirpByID(r.Context(), id)
		if err != nil {
			api.RespondWithError(w, http.StatusNotFound, "Chirp not found", err)
			return
		}
		api.RespondWithJSON(w, http.StatusOK, ChirpResponse{
			ID:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
			UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
			Body:      chirp.Body,
			UserID:    chirp.UserID.String(),
		})
	}
}
