package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/Vikuuu/Chirpy/internal/auth"
	"github.com/Vikuuu/Chirpy/internal/database"
)

type parameters struct {
	Body string `json:"body"`
	// UserID uuid.UUID `json:"user_id"`
}

type respBody struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	jwtToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error getting jwtToken: %s", err)
		respondWithError(w, 401, "Unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(jwtToken, cfg.secret)
	if err != nil {
		log.Printf("error validating token: %s", err)
		respondWithError(w, 401, "Unauthorized")
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := parameters{}
	err = decoder.Decode(&payload)
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
		UserID: userID,
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
	chirpID, err := uuid.Parse(pat)
	if err != nil {
		log.Fatalf("error parsing the id: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dat, err := cfg.db.GetChirp(context.Background(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, 404, "Not Found")
			return
		} else {
			log.Fatalf("error fetching the chirp: %s", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Fatalf("error parsing the id: %s", err)
		w.WriteHeader(500)
		return
	}

	jwtToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error getting jwtToken: %s", err)
		respondWithError(w, 401, "Unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(jwtToken, cfg.secret)
	if err != nil {
		log.Printf("error validating token: %s", err)
		respondWithError(w, 401, "Unauthorized")
		return
	}

	// check if the author of chirp and the logged in user are same?
	chirp, err := cfg.db.GetChirp(context.Background(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, 404, "Not Found")
			return
		} else {
			log.Fatalf("error getting chirp: %s", err)
			w.WriteHeader(500)
			return
		}
	}

	if userID != chirp.UserID {
		respondWithError(w, 403, "You don't have the permission to delete  this chirp")
		return
	}

	// if the user is the chirp author
	err = cfg.db.DeleteChirp(context.Background(), database.DeleteChirpParams{
		UserID: userID,
		ID:     chirpID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, 404, "chirp does not exist")
			return
		} else {
			log.Fatalf("error deleting chirp: %s", err)
			w.WriteHeader(500)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
