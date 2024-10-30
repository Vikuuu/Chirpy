package main

import "net/http"

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("Content-Type", "text/plain: charset=utf-8")
	w.WriteHeader(http.StatusOK)
	str := "OK"
	w.Write([]byte(str))
}
