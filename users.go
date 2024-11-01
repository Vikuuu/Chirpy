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

type userParameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type response struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type refreshResponse struct {
	Token string `json:"token"`
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
		ID:          dat.ID,
		CreatedAt:   dat.CreatedAt,
		UpdatedAt:   dat.UpdatedAt,
		Email:       dat.Email,
		IsChirpyRed: dat.IsChirpyRed,
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

	type loginParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type loginResponse struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	// decoding the input json
	decoder := json.NewDecoder(r.Body)
	params := loginParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	// getting user and checking password
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

	// if everything goes well, then create jwt

	expiresIn := time.Hour
	tokenSecret := cfg.secret

	jwtToken, err := auth.MakeJWT(dat.ID, tokenSecret, expiresIn)
	if err != nil {
		log.Fatalf("Error creating JWT token: %s", err)
		w.WriteHeader(500)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Fatalf("Error creating refresh token: %s", err)
		w.WriteHeader(500)
		return
	}
	// add created refresh token in the database
	err = cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: dat.ID,
	})
	if err != nil {
		log.Fatalf("Error adding refresh token to database: %s", err)
		w.WriteHeader(500)
		return
	}

	// return the reponse json
	data, err := json.Marshal(loginResponse{
		ID:           dat.ID,
		CreatedAt:    dat.CreatedAt,
		UpdatedAt:    dat.UpdatedAt,
		Email:        dat.Email,
		IsChirpyRed:  dat.IsChirpyRed,
		Token:        jwtToken,
		RefreshToken: refreshToken,
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

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Fatalf("error getting refresh token: %s", err)
		respondWithError(w, 401, "Refresh token not provided")
	}

	refreshUser, err := cfg.db.GetUserFromRefreshToken(context.Background(), refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, 401, "not a valid refresh token")
		} else {
			log.Fatalf("error retrieving refresh user: %s", err)
			w.WriteHeader(500)
			return
		}
	}

	if refreshUser.ExpiresAt.Before(time.Now()) || refreshUser.RevokedAt.Valid {
		respondWithError(w, 401, "refresh token expired")
		return
	}

	accessToken, err := auth.MakeJWT(refreshUser.UserID, cfg.secret, time.Hour)
	if err != nil {
		log.Fatalf("error creating access token: %s", err)
		w.WriteHeader(500)
		return
	}

	data, err := json.Marshal(refreshResponse{
		Token: accessToken,
	})
	if err != nil {
		log.Fatalf("error marshaling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Fatalf("error getting refresh token: %s", err)
		respondWithError(w, 401, "Refresh token not provided")
	}

	err = cfg.db.RevokeRefreshToken(context.Background(), database.RevokeRefreshTokenParams{
		RevokedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
		Token:     refreshToken,
	})
	if err != nil {
		log.Fatalf("error revoking token: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type updateParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateResponse struct {
	Email string `json:"email"`
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
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
	payload := updateParams{}
	err = decoder.Decode(&payload)
	if err != nil {
		log.Fatalf("error decoding JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	hashPsswd, err := auth.HashPassword(payload.Password)
	if err != nil {
		log.Fatalf("error hashing password: %s", err)
		w.WriteHeader(500)
		return
	}

	updatedEmail, err := cfg.db.EditUser(context.Background(), database.EditUserParams{
		Email:          payload.Email,
		HashedPassword: hashPsswd,
		UpdatedAt:      time.Now().UTC(),
		ID:             userID,
	})
	if err != nil {
		log.Fatalf("error updating user: %s", err)
		w.WriteHeader(500)
		return
	}

	data, err := json.Marshal(updateResponse{Email: updatedEmail})
	if err != nil {
		log.Fatalf("error marshaling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
