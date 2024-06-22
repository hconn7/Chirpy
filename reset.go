package main

import "net/http"

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Resetting"))
	cfg.fileServerHits = 0
}
