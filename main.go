package main

import (
	"net/http"
)

func main() {
	ServeMux := http.NewServeMux()
	Server := http.Server{
		Addr:    ":8080",
		Handler: ServeMux,
	}

	Server.ListenAndServe()
}
