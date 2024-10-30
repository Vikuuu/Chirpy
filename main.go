package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux.Handle(
		"/app/",
		http.StripPrefix(
			"/app",
			apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot))),
		),
	)
	mux.HandleFunc("GET  /api/healthz", handlerHealth)
	mux.HandleFunc("GET  /api/metrics", apiCfg.handlerMetric)
	mux.HandleFunc("POST /api/reset", apiCfg.handlerReset)

	log.Printf("Serving file from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) handlerMetric(w http.ResponseWriter, r *http.Request) {
	str := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
	r.Header.Add("Content-Type", "text/plain: charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(str))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
