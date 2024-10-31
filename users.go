package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/Vikuuu/Chirpy/internal/auth"
	"github.com/Vikuuu/Chirpy/internal/database"
)

type userParameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type response struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (apiCfg *apiConfig) handlerUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := userParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	hashPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Fatalf("Error hashing password: %s", err)
		w.WriteHeader(500)
		return
	}

	dat, err := apiCfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashPassword,
	})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(500)
		return
	}
	res := response{
		ID:        dat.ID,
		CreatedAt: dat.CreatedAt,
		UpdatedAt: dat.UpdatedAt,
		Email:     dat.Email,
	}

	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshaling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	const unauthMsg = "incorrect email or password"

	decoder := json.NewDecoder(r.Body)
	params := userParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	dat, err := cfg.db.GetUser(context.Background(), params.Email)
	if err != nil {
		respondWithError(w, 401, unauthMsg)
		return
	}

	err = auth.CheckPasswordHash(params.Password, dat.HashedPassword)
	if err != nil {
		respondWithError(w, 401, unauthMsg)
		return
	}

	data, err := json.Marshal(response{
		ID:        dat.ID,
		CreatedAt: dat.CreatedAt,
		UpdatedAt: dat.UpdatedAt,
		Email:     dat.Email,
	})
	if err != nil {
		log.Fatalf("Error marshaling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
