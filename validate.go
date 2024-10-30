package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

type parameters struct {
	Body string `json:"body"`
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	if len(params.Body) > 140 {
		errResp := "chirp is too long"
		respondWithError(w, 400, errResp)
		return
	}

	respondWithJSON(w, 200, params)
}

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
	w.WriteHeader(400)
	w.Write(errData)
}

func respondWithJSON(w http.ResponseWriter, code int, payload parameters) {
	type respBody struct {
		CleanedBody string `json:"cleaned_body"`
	}

	res := badWordReplacement(payload.Body)

	resBody := respBody{
		CleanedBody: res,
	}
	dat, err := json.Marshal(resBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func badWordReplacement(msg string) string {
	badWordList := []string{"kerfuffle", "sharbert", "fornax"}
	// lowerMsg := strings.ToLower(msg)
	msgList := strings.Split(msg, " ")

	for index, msg := range msgList {
		if slices.Contains(badWordList, strings.ToLower(msg)) {
			msgList[index] = "****"
		}
	}

	resMsg := strings.Join(msgList, " ")
	return resMsg
}
