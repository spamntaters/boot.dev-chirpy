package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spamntaters/boot.dev-chirpy/internal/api"
	"github.com/spamntaters/boot.dev-chirpy/internal/auth"
	"github.com/spamntaters/boot.dev-chirpy/internal/database"
)

type UserResponse struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Email        string `json:"email"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type RefreshTokenResponse struct {
	Token string `json:"token"`
}

func HandleCreateUser(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Password string `json:"password"`
			Email    string `json:"email"`
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			api.RespondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			api.RespondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
		processedParams := database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
		}
		data, err := cfg.DB.CreateUser(r.Context(), processedParams)
		if err != nil {
			api.RespondWithError(w, http.StatusBadRequest, "Email is required and must be unique", err)
			return
		}
		user := UserResponse{
			ID:        data.ID.String(),
			CreatedAt: data.CreatedAt.Format(time.RFC3339),
			UpdatedAt: data.UpdatedAt.Format(time.RFC3339),
			Email:     data.Email,
		}
		api.RespondWithJSON(w, http.StatusCreated, user)
	}
}

func HandleLogin(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			api.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}
		expireDuration := 1 * time.Hour
		data, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
		if err != nil {
			api.RespondWithError(w, http.StatusNotFound, "User not found", err)
			return
		}
		if err := auth.CheckPasswordHash(params.Password, data.HashedPassword); err != nil {
			api.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials", nil)
			return
		}
		token, err := auth.MakeJWT(data.ID, cfg.Secret, expireDuration)
		if err != nil {
			api.RespondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}

		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			api.RespondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}

		cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
			Token:  refreshToken,
			UserID: data.ID,
		})

		user := UserResponse{
			ID:           data.ID.String(),
			CreatedAt:    data.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    data.UpdatedAt.Format(time.RFC3339),
			Email:        data.Email,
			Token:        token,
			RefreshToken: refreshToken,
		}
		api.RespondWithJSON(w, http.StatusOK, user)
	}
}

func HandleResetUsers(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Platform != "dev" {
			api.RespondWithError(w, http.StatusForbidden, "Reset only available in dev environments", nil)
			return
		}
		err := cfg.DB.ResetUsers(r.Context())
		if err != nil {
			api.RespondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleRefreshToken(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			api.RespondWithError(w, http.StatusUnauthorized, "Invlaid Refresh token", err)
			return
		}
		userID, err := cfg.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
		if err != nil {
			api.RespondWithError(w, http.StatusUnauthorized, "Invalid Refresh token", err)
			return
		}

		expireDuration := 1 * time.Hour

		token, err := auth.MakeJWT(userID, cfg.Secret, expireDuration)
		if err != nil {
			api.RespondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}

		api.RespondWithJSON(w, http.StatusOK, RefreshTokenResponse{
			Token: token,
		})
	}
}

func HandleRevokeToken(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			api.RespondWithError(w, http.StatusUnauthorized, "Invlaid Refresh token", err)
			return
		}
		cfg.DB.RevokeRefreshToken(r.Context(), refreshToken)
		api.RespondWithJSON(w, http.StatusNoContent, nil)
	}
}
