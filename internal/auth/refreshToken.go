package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return "", errors.New("Couldn't make refresh token")
	}
	refreshToken := hex.EncodeToString(key)
	return refreshToken, nil
}
