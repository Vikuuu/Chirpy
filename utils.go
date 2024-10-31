package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type respError struct {
		Error string `json:"error"`
	}
	respErr := respError{
		Error: msg,
	}
	errData, err := json.Marshal(respErr)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(errData)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	var dat []byte
	var err error

	switch v := payload.(type) {
	case respBody:
		// Single respBody case
		dat, err = json.Marshal(v)

	case []respBody:
		// Slice of respBody structs
		dat, err = json.Marshal(v)

	default:
		// Unsupported type
		log.Printf("Invalid Payload type")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}
