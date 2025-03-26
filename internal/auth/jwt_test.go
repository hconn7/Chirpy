package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMakeJWT(t *testing.T) {
	// Setup
	userID := uuid.New()
	tokenSecret := "secretkey"
	expiresIn := 1 * time.Hour

	// Test: Create JWT
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Log the token for debugging purposes
	t.Log("Generated token:", tokenString)

	// Test: Try to parse the token and validate the claims
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure the token is signed using the correct signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("Expected signing method HMAC, got %v", token.Method)
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		t.Fatalf("Error parsing token: %v", err)
	}

	// Ensure parsedToken is not nil
	if parsedToken == nil {
		t.Fatal("Parsed token is nil")
	}

	// Extract claims and check the subject (userID)
	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		t.Fatalf("Failed to cast claims to RegisteredClaims")
	}
	assert.Equal(t, userID.String(), claims.Subject)
}
