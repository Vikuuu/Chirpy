package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/Vikuuu/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	secret         string
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Connection cannot be made to DB")
	}
	defer db.Close()

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		secret:         os.Getenv("SECRET"),
	}

	mux.Handle(
		"/app/",
		http.StripPrefix(
			"/app",
			apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot))),
		),
	)
	mux.HandleFunc("GET  /api/healthz", handlerHealth)
	mux.HandleFunc("GET  /admin/metrics", apiCfg.handlerMetric)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerPostChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUser)
	mux.HandleFunc("GET  /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET  /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	log.Printf("Serving file from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
