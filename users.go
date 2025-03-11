package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/spamntaters/boot.dev-chirpy/internal/auth"
	"github.com/spamntaters/boot.dev-chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token,omitempty"`
}

func (cfg *apiConfig) addUserHandler(resWriter http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string
		Email    string
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(resWriter, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	processedParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}
	if err != nil {
		respondWithError(resWriter, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	data, err := cfg.db.CreateUser(req.Context(), processedParams)
	if err != nil {
		respondWithError(resWriter, http.StatusBadRequest, "Email is required and must be unique", err)
		return
	}

	user := User{
		ID:        data.ID,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.CreatedAt,
		Email:     data.Email,
	}

	respondWithJSON(resWriter, http.StatusCreated, user)
}

func (cfg *apiConfig) resetUsersHandler(resWriter http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(resWriter, http.StatusForbidden, "Rest only available in dev environments", nil)
		return
	}
	err := cfg.db.ResetUsers(req.Context())
	if err != nil {
		respondWithError(resWriter, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	resWriter.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handleUserLogin(resWriter http.ResponseWriter, req *http.Request) {
	type parameters = struct {
		Email            string
		Password         string
		ExpiresInSeconds int64
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	decoder.Decode(&params)

	expireDuration := 1 * time.Hour

	if params.ExpiresInSeconds != 0 {
		expireDuration = time.Second * time.Duration(params.ExpiresInSeconds)
	}

	data, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		respondWithError(resWriter, http.StatusNotFound, "User not found", err)
		return
	}
	noAuth := auth.CheckPasswordHash(params.Password, data.HashedPassword)
	if noAuth != nil {
		respondWithError(resWriter, http.StatusUnauthorized, "User is not authorized", noAuth)
		return
	}
	token, err := auth.MakeJWT(data.ID, cfg.secret, expireDuration)
	if err != nil {
		respondWithError(resWriter, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	user := User{
		ID:        data.ID,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
		Email:     data.Email,
		Token:     token,
	}
	respondWithJSON(resWriter, http.StatusOK, user)
}
