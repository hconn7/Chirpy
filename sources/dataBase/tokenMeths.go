package database

import (
	"errors"
	"time"
)

func (db *DB) SaveRefreshToken(userID int, token string) error {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return err
	}

	refreshToken := RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	dbStructure.RefreshTokens[token] = refreshToken

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) RevokeRefreshToken(token string) error {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return err
	}
	delete(dbStructure.RefreshTokens, token)
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}
func (db *DB) UserForRefreshToken(token string) (UserDat, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return UserDat{}, err
	}

	refreshToken, ok := dbStructure.RefreshTokens[token]
	if !ok {
		return UserDat{}, errors.New("Cannot find token to user")
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return UserDat{}, errors.New("Cannot find token to user")
	}

	user, err := db.GetUser(refreshToken.UserID)
	if err != nil {
		return UserDat{}, err
	}

	return user, nil
}
func (db *DB) GetUser(id int) (UserDat, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return UserDat{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return UserDat{}, errors.New("Cannot find user")

	}

	return user, nil
}
