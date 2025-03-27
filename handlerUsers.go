package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hconn7/Chirpy/internal/auth"
	"github.com/hconn7/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type Response struct {
		Id         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Email      string    `json:"email"`
		Password   string    `json:"password"`
		Sub        bool      `json:"is_chirpy_red"`
	}
	params := Params{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode params", err)
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "Internal Hashing error", err)
	}

	userEmail := params.Email
	user, error := cfg.dbQueries.CreateUser(context.Background(), database.CreateUserParams{
		Email:          userEmail,
		HashedPassword: hashedPassword})

	if err != nil {
		log.Fatal("user not created %s", error)
	}
	respondWithJson(w, 201, Response{
		Id:         user.ID,
		Created_at: user.CreatedAt,
		Updated_at: user.UpdatedAt,
		Email:      user.Email,
		Password:   params.Password,
		Sub:        user.IsChirpyRed,
	})
}

func (cfg *apiConfig) handlerValidateLogin(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var params Params
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Internal json reading error", err)
	}
	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "No user exists, please check email or password", err)
	}

	if err := auth.CheckPasswordHash(user.HashedPassword, params.Password); err != nil {
		respondWithError(w, 401, "Password or email is incorrect", err)
	}

	// Refresh Token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 500, "Couldn't make refresh token", err)
	}
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID})
	if err != nil {
		respondWithError(w, 500, "Couldn't create and retrive token", err)
	}

	//JWT
	token, err := auth.MakeJWT(user.ID, cfg.JwtSecret, time.Duration(time.Hour))
	if err != nil {
		respondWithError(w, 500, "Couldn't make JWT", err)
	}

	type Response struct {
		Id           uuid.UUID `json:"id"`
		Created_at   time.Time `json:"created_at"`
		Updated_at   time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		Password     string    `json:"password"`
		Sub          bool      `json:"is_chirpy_red"`
	}

	respondWithJson(w, 200, Response{
		Id:           user.ID,
		Created_at:   user.CreatedAt,
		Updated_at:   user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		Sub:          user.IsChirpyRed,
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

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Token    string `json:"token"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "No Auth header", err)
	}

	var params Params
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, 401, "Missing data", err)
	}
	fmt.Printf("Users email: %v, User password: %v", params.Email, params.Password)
	userID, err := auth.ValidateJWT(authToken, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, 401, "Recieved Token but couldn't validate it", err)
	}

	user, err := cfg.dbQueries.GetUserByID(r.Context(), userID)

	password, _ := auth.HashPassword(params.Password)
	cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             user.ID,
		Email:          params.Email,
		HashedPassword: password})
	type Response struct {
		Email string `json:"email"`
	}
	respondWithJson(w, 200, Response{Email: params.Email})
}
