package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
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

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
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
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	user, err := cfg.DB.LookupByEmail(reqBody.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqBody.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	expirationTime := time.Hour * 24
	if reqBody.ExpiresInSeconds != nil {
		expires := time.Duration(*reqBody.ExpiresInSeconds) * time.Second
		if expires < time.Hour*24 {
			expirationTime = expires
		}
	}

	claims := &jwt.StandardClaims{
		Issuer:    "chirpy",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(expirationTime).Unix(),
		Subject:   strconv.Itoa(user.ID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":    strconv.Itoa(user.ID),
		"email": reqBody.Email,
		"token": tokenString,
	})
}
func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleUpdateUser called")

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
	fmt.Println("Token extracted:", tokenStr)

	// Parse the token with claims
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		fmt.Println("Token parsing error:", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || claims.Subject == "" {
		fmt.Println("Invalid token claims")
		respondWithError(w, http.StatusUnauthorized, "Invalid token claims")
		return
	}
	fmt.Println("Token claims:", claims)

	// Decode request body to get new user email and password
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	fmt.Println("Request body decoded:", reqBody)

	// Convert the subject claim (user ID) to an integer
	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}
	fmt.Println("User ID extracted:", userId)

	// Update the user in the database
	fmt.Println("Attempting to update user")
	updatedUser, err := cfg.DB.UpdateUser(userId, reqBody.Email, reqBody.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user")
		return
	}

	fmt.Println("user Upated")

	// Construct and send the response
	response := map[string]interface{}{
		"id":    updatedUser.ID,
		"email": updatedUser.Email,
	}
	fmt.Println("attempting to respond")
	respondWithJSON(w, http.StatusOK, response)
}
