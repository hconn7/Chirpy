package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hconn7/Chirpy/internal/auth"
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
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "No token", err)
	}

	userID, err := auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, 401, "Token not validated", err)
	}

	newChirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   deProfane,
		UserID: userID,
	})
	fmt.Println("Chirp Created\n", newChirp.Body, newChirp.ID)
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
		UserID:    userID,
	})
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("author_id")
	sortHeader := r.URL.Query().Get("sort")
	chirpsSlice, err := cfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Error retreiving Chirps", err)
	}
	if s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			fmt.Println("Error parsing UUID:", err)
			return
		}
		chirps, err := cfg.dbQueries.GetChirpByUserID(r.Context(), id)
		if err != nil {
			respondWithError(w, 401, "user doesn't exist", err)
		}

		responeChirps := []Chirp{}
		for _, chirp := range chirps {
			singleChirp := Chirp{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserID:    chirp.UserID,
			}

			responeChirps = append(responeChirps, singleChirp)
		}
		respondWithJson(w, 200, responeChirps)
	} else {

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
		if sortHeader == "asc" || sortHeader == "" {
			sort.Slice(responseChirps, func(i, j int) bool { return responseChirps[i].CreatedAt.Before(responseChirps[j].CreatedAt) })

			respondWithJson(w, 200, responseChirps)
		} else {
			sort.Slice(responseChirps, func(i, j int) bool { return responseChirps[i].CreatedAt.After(responseChirps[j].CreatedAt) })
			respondWithJson(w, 200, responseChirps)

		}
	}
}

func (cfg *apiConfig) hanlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	if id == "" {
		respondWithError(w, 404, "no Id", errors.New("error"))
		fmt.Println("No ID in request")
		return
	}
	fmt.Println(id)
	chirpID, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, 404, "No chirp", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "incorrect id or chirp doesn't exist", err)
		return
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	if id == "" {
		fmt.Println("No ID in request")
	}
	fmt.Println(id)
	chirpID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "No auth header", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, 401, "JWT invalid", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "No chirp found", err)
		fmt.Println("Couln't find chirp with id:", chirpID)
		return
	}

	if userID != chirp.UserID {
		respondWithError(w, 403, "Not authorized to delete this chirp", err)
		return
	}

	if err := cfg.dbQueries.DeleteChirp(r.Context(), chirpID); err != nil {
		respondWithError(w, 404, "issue deleting chirp", err)
		return
	}
	respondWithJson(w, 204, "")

}
