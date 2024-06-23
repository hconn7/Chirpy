package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/hconn7/Chirpy/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handler Triggered")
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type UserDat struct {
		ID         int    `json:"id"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		Expiration string `json:"expires_in_seconds"`
	}

	fmt.Printf("Received %s request for %s\n", r.Method, r.URL.Path)
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not decode")
		return
	}

	userDat, err := cfg.DB.CreateUser(params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, UserDat{
		ID:    userDat.ID,
		Email: userDat.Email,
	})
}

// Congirm login
func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type UserDat struct {
		ID           int    `json:"id"`
		Email        string `json:"email"`
		Password     string `json:"password"`
		Expiration   string `json:"expires_in_seconds"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not decode")
		return
	}

	user, err := cfg.DB.LookupByEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Email or password incorrect")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.Password))

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Email or password incorrect")
		return
	}
	respondWithJSON(w, http.StatusOK, UserDat{
		ID:    user.ID,
		Email: user.Email,
	})
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}
	type UserDat struct {
		ID           int    `json:"id"`
		Email        string `json:"email"`
		Password     string `json:"password"`
		Expiration   string `json:"expires_in_seconds"`
		RefreshToken string `json:"refresh_token"`
		Token        string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	fmt.Println("Attmepting to authorize")
	// Validate user credentials
	user, err := cfg.DB.LookupByEmail(reqBody.Email)
	if err != nil {
		fmt.Println("Error in LookupByEmail", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	fmt.Println("authorize complete")

	// Set default expiration to 24 hours
	expirationTime := time.Hour * 24
	if reqBody.ExpiresInSeconds != nil {
		expires := time.Duration(*reqBody.ExpiresInSeconds) * time.Second
		if expires < time.Hour*24 {
			expirationTime = expires
		}
	}

	// Create JWT claims
	claims := &jwt.StandardClaims{
		Issuer:    "chirpy",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(expirationTime).Unix(),
		Subject:   strconv.Itoa(user.ID), // Convert the user ID to a string
	}

	// Create the token with the specified claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}
	refreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token")
		return
	}
	err = cfg.DB.SaveRefreshToken(user.ID, refreshToken)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token")
		return
	}
	respondWithJSON(w, http.StatusOK, UserDat{
		ID:           user.ID,
		Email:        user.Email,
		Token:        tokenString,
		RefreshToken: refreshToken,
	})
}
func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {

	// Extract and validate the JWT from the request header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Missing Authorization header")
		return
	}

	// Strip "Bearer " prefix from the token
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader {
		respondWithError(w, http.StatusUnauthorized, "Invalid token format")
		return
	}

	// Parse the token with claims
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || claims.Subject == "" {
		respondWithError(w, http.StatusUnauthorized, "Invalid token claims")
		return
	}

	// Decode request body to get new user email and password
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}
	updatedUser, err := cfg.DB.UpdateUser(userId, reqBody.Email, reqBody.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user")
		return
	}

	response := map[string]interface{}{
		"id":    updatedUser.ID,
		"email": updatedUser.Email,
	}
	respondWithJSON(w, http.StatusOK, response)
}
