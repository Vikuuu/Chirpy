package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/Vikuuu/Chirpy/internal/auth"
)

type polkaParams struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("error getting Api Key: %s", err)
		respondWithError(w, 401, "Unauthorized")
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := polkaParams{}
	err = decoder.Decode(&payload)
	if err != nil {
		log.Fatalf("error decoding JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	if payload.Event != "user.upgraded" {
		respondWithError(w, 204, "No Content")
		return
	}

	err = cfg.db.UpgradeUserToRed(context.Background(), payload.Data.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, 404, "User Not Found")
			return
		} else {
			log.Fatalf("error updating user: %s", err)
			w.WriteHeader(500)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
