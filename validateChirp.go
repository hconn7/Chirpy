package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Response struct {
	CleanedBody string `json:"cleaned_body"`
}
type Chirp struct {
	Body string `json:"body"`
}
type ErrorResponse struct {
	Error error `json:"error"`
}

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirp := Chirp{}
	err := decoder.Decode(&chirp)
	if err != nil {

		w.WriteHeader(500)

	}
	deProfane := CheckProfanityChirp(chirp.Body)
	if len(chirp.Body) >= 140 {
		respondWithError(w, 400, "Chirp length is too long", err)
	}
	response := Response{CleanedBody: deProfane}
	fmt.Print(response)
	payload, err := json.Marshal(response)
	fmt.Print(payload)
	if err != nil {

		respondWithError(w, 400, "Error Marshaling Response", err)

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(payload))
}

func CheckProfanityChirp(chirps ...string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	var cleanedChirps []string

	for _, chirp := range chirps {
		words := strings.Split(chirp, " ")

		for i, word := range words {
			loweredWord := strings.ToLower(word)

			for _, badWord := range badWords {
				if loweredWord == badWord {
					words[i] = "****"
				}
			}
		}

		cleanedChirps = append(cleanedChirps, strings.Join(words, " "))
	}

	return strings.Join(cleanedChirps, " ")
}
