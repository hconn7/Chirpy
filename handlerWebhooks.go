package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/hconn7/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		UserID uuid.UUID `json:"user_id"`
	}
	type Params struct {
		Event string `json:"event"`
		Data  Data   `json:"data"`
	}
	header, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, 401, "Header not included", err)
	}
	if header != cfg.ApiKey {
		respondWithError(w, 401, "Wrong api key!", errors.New(""))
	}

	var params Params
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, 401, "Couldn't decode params", err)
	}

	if params.Event != "user.upgraded" {
		respondWithError(w, 204, "Wrong event", errors.New("Wrong event"))
	}

	if err := cfg.dbQueries.UpdateChirpyRed(r.Context(), params.Data.UserID); err != nil {
		respondWithError(w, 404, "No user found with ID", err)
	}
	respondWithJson(w, 204, "")
}
