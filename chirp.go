package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/Vikuuu/Chirpy/internal/database"
)

type parameters struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type respBody struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	payload := parameters{}
	err := decoder.Decode(&payload)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	if len(payload.Body) > 140 {
		errResp := "chirp is too long"
		respondWithError(w, 400, errResp)
		return
	}
	payload.Body = badWordReplacement(payload.Body)

	dat, err := cfg.db.CreateChirpForUser(context.Background(), database.CreateChirpForUserParams{
		Body:   payload.Body,
		UserID: payload.UserID,
	})
	if err != nil {
		log.Fatalf("Error creating chirp: %s", err)
		w.WriteHeader(500)
		return
	}

	respPayload := respBody{
		ID:        dat.ID,
		CreatedAt: dat.CreatedAt,
		UpdatedAt: dat.UpdatedAt,
		Body:      dat.Body,
		UserID:    dat.UserID,
	}

	respondWithJSON(w, 201, respPayload)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dat, err := cfg.db.GetChirps(context.Background())
	if err != nil {
		log.Fatalf("Error creating chirp: %s", err)
		w.WriteHeader(500)
		return
	}
	var resp []respBody
	for _, chirp := range dat {
		i := respBody{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		resp = append(resp, i)
	}

	respondWithJSON(w, 200, resp)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	pat := r.PathValue("chirpID")
	log.Printf(pat)
	log.Printf(r.Pattern)
	chirpID, err := uuid.Parse(pat)
	if err != nil {
		log.Fatalf("error parsing the id: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dat, err := cfg.db.GetChirp(context.Background(), chirpID)
	if err != nil {
		log.Fatalf("error fetching the chirp: %s", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	respPayload := respBody{
		ID:        dat.ID,
		CreatedAt: dat.CreatedAt,
		UpdatedAt: dat.UpdatedAt,
		Body:      dat.Body,
		UserID:    dat.UserID,
	}

	respondWithJSON(w, 200, respPayload)
}
