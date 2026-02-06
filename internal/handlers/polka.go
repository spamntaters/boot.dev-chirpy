package handlers

import (
	"encoding/json"
	"net/http"
	
	"github.com/google/uuid"
	"github.com/spamntaters/boot.dev-chirpy/internal/api"
)

type EventInput struct {
	Event string `json:"event"`
	Data struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func HandlePolkaEvent(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		eventParams := EventInput{}
		err := decoder.Decode(&eventParams)
		if err != nil {
			api.RespondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
		
		if eventParams.Event != "user.upgraded" {
			api.RespondWithJSON(w, http.StatusNoContent, nil)
			return
		}
		
		err = cfg.DB.UpgradeUserByID(r.Context(), eventParams.Data.UserID)
		if err != nil {
			api.RespondWithError(w, http.StatusNotFound, "User not found", err)
			return
		}

		api.RespondWithJSON(w, http.StatusNoContent, nil)
	}
}

