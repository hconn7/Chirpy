package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hconn7/Chirpy/internal/database"
)

type Request struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}
type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}
type ErrorResponse struct {
	Error error `json:"error"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	request := Request{}
	err := decoder.Decode(&request)
	fmt.Printf("Request body: %+v\n", request)
	fmt.Printf("User ID Type: %T, Value: %s\n", request.UserID, request.UserID)

	if err != nil {

		w.WriteHeader(500)

	}
	deProfane := CheckProfanityChirp(request.Body)

	if len(request.Body) >= 140 {
		respondWithError(w, 400, "Chirp length is too long", err)
	}
	if err != nil {

		respondWithError(w, 400, "Error Marshaling Response", err)

	}

	user, err := cfg.dbQueries.GetUserByID(r.Context(), request.UserID)
	if err != nil {
		// User doesn't exist
		fmt.Printf("User %s not found: %v\n", request.UserID, err)
		respondWithError(w, 404, "User not found", err)
		return
	}
	newChirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   deProfane,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, 500, "Error creating chirp", err)
		return
	}

	fmt.Println(newChirp.Body)

	respondWithJson(w, 201, Chirp{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    request.UserID,
	})
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirpsSlice, err := cfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Error retreiving Chirps", err)
	}
	responseChirps := []Chirp{}
	for _, chirp := range chirpsSlice {
		singleChirp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}

		responseChirps = append(responseChirps, singleChirp)
	}

	respondWithJson(w, 200, responseChirps)
}
func (cfg *apiConfig) hanlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpsID")
	if id == "" {
		fmt.Println("No ID in request")
	}
	fmt.Println(id)
	chirpID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Fatal("Chirp not found with the given id: ", chirpID)
	}

	respondWithJson(w, 200, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.ID,
	})

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
