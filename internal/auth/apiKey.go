package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("No Auth Header")
	}

	const AuthPrefix = "ApiKey "
	if !strings.HasPrefix(authHeader, AuthPrefix) {
		return "", errors.New("Invalid Auth format")
	}

	return strings.TrimPrefix(authHeader, AuthPrefix), nil
}
