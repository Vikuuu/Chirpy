package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.Handle(
		"/app/",
		http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))),
	)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add("Content-Type", "text/plain: charset=utf-8")
		w.WriteHeader(http.StatusOK)
		str := "OK"
		w.Write([]byte(str))
	})

	log.Printf("Serving file from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
