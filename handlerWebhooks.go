package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func (cfg *apiConfig) handlerUpgrade(w http.ResponseWriter, r *http.Request) {
	type DataPayload struct {
		UserID int `json:"user_id"`
	}
	type parameters struct {
		Event string      `json:"event"`
		Data  DataPayload `json:"data"`
	}
	err := cfg.apiAuthorization(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Faulty api key")
		return
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Request not applicable")
		return
	}
	user, err := cfg.DB.GetUser(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Perform subscription upgrade (set Subscription to true)
	_, err = cfg.DB.UserUpgrade(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error confirming subscription")
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
func (cfg *apiConfig) apiAuthorization(headers http.Header) error {
	apiHeader := headers.Get("Authorization")
	if apiHeader == "" {
		return errors.New("no header")
	}
	splitAuth := strings.Split(apiHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return errors.New("no authorization")
	}
	if splitAuth[1] != cfg.apiKey {
		return errors.New("missing key")
	}

	return nil
}
