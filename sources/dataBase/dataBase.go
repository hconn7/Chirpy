package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var ErrChirpNotFound = errors.New("chirp not found")

type DB struct {
	path string
	mu   *sync.RWMutex
}
type DBStructure struct {
	Users  map[int]UserDat `json:"users"`
	Chirps map[int]Chirp   `json:"chirps"`
}

type UserDat struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:   id,
		Body: body,
	}
	dbStructure.Chirps[id] = chirp
	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}
func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Users:  map[int]UserDat{},
		Chirps: map[int]Chirp{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}
	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	dbStructure := DBStructure{}
	dat, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}
	err = json.Unmarshal(dat, &dbStructure)
	if err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dat, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp, ok := dbStructure.Chirps[id]
	if !ok {
		return Chirp{}, ErrChirpNotFound
	}

	return chirp, nil
}

func (db *DB) CreateUser(email, password string) (UserDat, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return UserDat{}, nil
	}
	for _, user := range dbStruct.Users {
		if user.Email == email {
			return UserDat{}, errors.New("Email is alredy in use")
		}
	}
	cryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return UserDat{}, err
	}
	id := len(dbStruct.Users) + 1
	userDat := UserDat{
		ID:       id,
		Email:    email,
		Password: string(cryptedPassword),
	}
	dbStruct.Users[id] = userDat
	err = db.writeDB(dbStruct)
	if err != nil {
		return UserDat{}, err
	}
	return userDat, nil
}

func (db *DB) LookupByEmail(email string) (UserDat, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return UserDat{}, nil
	}

	for _, user := range dbStruct.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return UserDat{}, errors.New("user not found")
}

func (db *DB) UpdateUser(userID int, newEmail, newPassword string) (*UserDat, error) {
	fmt.Println("UpdateUser called with ID:", userID, "Email:", newEmail, "Password:", newPassword)

	dbStruct, err := db.loadDB()
	if err != nil {
		fmt.Println("Error loading DB:", err)
		return nil, err
	}

	user, ok := dbStruct.Users[userID]
	if !ok {
		fmt.Println("User ID not found in DB")
		return nil, errors.New("user not found")
	}

	for id, u := range dbStruct.Users {
		if u.Email == newEmail && id != userID {
			fmt.Println("Email already in use by another user")
			return nil, errors.New("email is already in use")
		}
	}

	cryptedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error encrypting password:", err)
		return nil, err
	}

	user.Email = newEmail
	user.Password = string(cryptedPassword)
	dbStruct.Users[userID] = user

	err = db.writeDB(dbStruct)
	if err != nil {
		fmt.Println("Error writing DB:", err)
		return nil, err
	}

	updatedUser := user
	updatedUser.Password = ""
	fmt.Println("User updated successfully:", updatedUser)
	return &updatedUser, nil
}
