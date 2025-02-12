package main

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/hconn7/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")
	if s != "" {
		authorID, err := strconv.Atoi(s)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Issues converting author_id to integer")
			return
		}

		chirps, err := cfg.DB.GetChirpByAuth(authorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Issues finding chirps for author")
			return
		}

		respondWithJSON(w, http.StatusOK, chirps)
		return
	}

	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := make([]Chirp, len(dbChirps))
	for i, dbChirp := range dbChirps {
		chirps[i] = Chirp{
			ID:       dbChirp.ID,
			Body:     dbChirp.Body,
			AuthorID: dbChirp.AuthorID,
		}
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	respondWithJSON(w, http.StatusOK, chirps)

	switch sortOrder {
	case "desc":
		sort.Slice(dbChirps, func(i, j int) bool {
			return dbChirps[i].ID > dbChirps[j].ID
		})
	case "asc", "":
		fallthrough
	default:
		sort.Slice(dbChirps, func(i, j int) bool {
			return dbChirps[i].ID < dbChirps[j].ID
		})
	}

	chirps = []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:       dbChirp.ID,
			Body:     dbChirp.Body,
			AuthorID: dbChirp.AuthorID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}
func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := strconv.Atoi(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	dbChirp, err := cfg.DB.GetChirp(chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:   dbChirp.ID,
		Body: dbChirp.Body,
	})
}
func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	userIDStr, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't Validate JWT")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse user ID")
	}

	chirpIDString := r.PathValue("chirpID")
	chirpID, err := strconv.Atoi(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}
	if chirpID != userID {
		respondWithError(w, http.StatusForbidden, "Unathorized to delete")
	}

	if err := cfg.DB.DeleteChirp(chirpID, userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}
	respondWithJSON(w, http.StatusNoContent, "Chirp deleted!")
}
