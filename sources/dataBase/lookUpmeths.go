package database

import (
	"errors"
	"strings"
)

func (db *DB) LookupByRefreshToken(refreshToken string) (UserDat, error) {
	dbStruct, err := db.LoadDB() // Assuming LoadDB() is correctly implemented elsewhere
	if err != nil {
		return UserDat{}, err
	}

	for _, user := range dbStruct.Users { // Assuming dbStruct.Users is a list of UserDat
		if user.RefreshToken == refreshToken {
			return user, nil
		}
	}
	return UserDat{}, errors.New("User not found via Refresh Token")
}

func (db *DB) LookupByEmail(email string) (UserDat, error) {
	dbStruct, err := db.LoadDB()
	if err != nil {
		return UserDat{}, err
	}

	// Normalize email inputs for safer comparison
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	// Iterate through the map to find the email
	for _, user := range dbStruct.Users {
		if strings.ToLower(strings.TrimSpace(user.Email)) == normalizedEmail {
			return user, nil
		}
	}
	return UserDat{}, errors.New("user not found")
}
