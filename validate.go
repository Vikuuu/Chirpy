package main

import (
	"slices"
	"strings"
)

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
