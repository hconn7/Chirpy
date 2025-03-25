package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Email string `json:"email"`
	}

	type Response struct {
		Id         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Email      string    `json:"email"`
	}
	params := Params{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode params", err)
	}

	userEmail := params.Email
	user, error := cfg.dbQueries.CreateUser(context.Background(), userEmail)

	if err != nil {
		log.Fatal("user not created %s", error)
	}
	respondWithJson(w, 201, Response{
		Id:         user.ID,
		Created_at: user.CreatedAt,
		Updated_at: user.UpdatedAt,
		Email:      user.Email,
	})
}

func (cfg *apiConfig) handlerResetUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		respondWithJson(w, 403, "Forbidden")
	}
	if err := cfg.dbQueries.DeleteUsers(context.Background()); err != nil {
		log.Fatal("error deleting users %v")
	}

	respondWithJson(w, 200, "Deleted users")
}
