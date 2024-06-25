package database

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) CreateUser(email, password string) (UserDat, error) {
	dbStruct, err := db.LoadDB()
	if err != nil {
		return UserDat{}, err
	}

	// Check if email is already in use
	for _, user := range dbStruct.Users {
		if user.Email == email {
			return UserDat{}, errors.New("Email is already in use")
		}
	}

	// Generate hashed password
	cryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return UserDat{}, err
	}

	// Calculate next available ID (simplified for practice)
	id := len(dbStruct.Users) + 1

	// Create new UserDat object
	userDat := UserDat{
		ID:           id,
		Email:        email,
		Password:     string(cryptedPassword),
		Subscription: false, // Use the provided subscription value
	}

	// Log the subscription status being set
	fmt.Printf("Setting subscription status for user %d: %v\n", userDat.ID, userDat.Email)

	// Update database struct
	dbStruct.Users[id] = userDat

	// Persist changes to the database
	err = db.writeDB(dbStruct)
	if err != nil {
		return UserDat{}, err
	}

	return userDat, nil
}

func (db *DB) UpdateUser(userID int, newEmail, newPassword string) (*UserDat, error) {

	dbStruct, err := db.LoadDB()
	if err != nil {
		return nil, err
	}

	user, ok := dbStruct.Users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}

	for id, u := range dbStruct.Users {
		if u.Email == newEmail && id != userID {
			return nil, errors.New("email is already in use")
		}
	}

	cryptedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Email = newEmail
	user.Password = string(cryptedPassword)
	dbStruct.Users[userID] = user

	err = db.writeDB(dbStruct)
	if err != nil {
		return nil, err
	}

	updatedUser := user
	updatedUser.Password = ""
	return &updatedUser, nil
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
func (db *DB) UserUpgrade(id int) (*UserDat, error) {
	dbStruct, err := db.LoadDB()
	if err != nil {
		return nil, err
	}

	user, ok := dbStruct.Users[id]
	if !ok {
		return nil, errors.New("Cannot find user")
	}

	user.Subscription = true

	dbStruct.Users[id] = user

	err = db.writeDB(dbStruct)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
