package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/hconn7/Chirpy/internal/auth"
	"github.com/hconn7/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerValidateRefreshToken(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Token string `json:"token"`
	}
	refreshTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Not Refresh Token\n", err)
	}
	tokenDB, err := cfg.dbQueries.GetRefreshTokenByToken(r.Context(), refreshTok)
	if err != nil {
		respondWithError(w, 401, "Token not in system\n", err)
	}
	if time.Now().After(tokenDB.ExpiresAt) || tokenDB.RevokedAt.Valid {
		respondWithError(w, 401, "Token Expired or revoked", nil)
	}
	accessToken, err := auth.MakeJWT(tokenDB.UserID, cfg.JwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, 500, "Error making Token\n", err)
	}
	respondWithJson(w, 200, Response{Token: accessToken})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Not Refresh Token", err)
	}
	tokenDB, err := cfg.dbQueries.GetRefreshTokenByToken(r.Context(), refreshTok)
	if err != nil {
		respondWithError(w, 401, "Token not in system", err)
	}
	err = cfg.dbQueries.UpdateRefreshToken(r.Context(), database.UpdateRefreshTokenParams{
		Token:     tokenDB.Token,
		UpdatedAt: time.Now(),
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	})
	respondWithJson(w, 204, "")
}
