package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) handlerUpgrade(w http.ResponseWriter, r *http.Request) {
	type DataPayload struct {
		UserID int `json:"user_id"`
	}
	type parameters struct {
		Event string      `json:"event"`
		Data  DataPayload `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
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
