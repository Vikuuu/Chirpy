package main

import (
	"context"
	"log"
	"net/http"
	"os"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	paltform := os.Getenv("PLATFORM")
	if paltform != "dev" {
		respondWithError(w, 403, "Platform in not DEV")
	}
	err := cfg.db.DeleteAllUsers(context.Background())
	if err != nil {
		log.Fatalf("Error deleting all the users: %s", err)
		w.WriteHeader(500)
		return
	}
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}
