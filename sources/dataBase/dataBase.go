package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
	"sync"
	"time"
)

var ErrChirpNotFound = errors.New("chirp not found")

type RefreshToken struct {
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
type DB struct {
	path string
	mu   *sync.RWMutex
}
type DBStructure struct {
	Users         map[int]UserDat         `json:"users"`
	Chirps        map[int]Chirp           `json:"chirps"`
	RefreshTokens map[string]RefreshToken `json:"refresh_token"`
}

type UserDat struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
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
	dbStructure, err := db.LoadDB()
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
	dbStructure, err := db.LoadDB()
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
		Users:         map[int]UserDat{},
		Chirps:        map[int]Chirp{},
		RefreshTokens: map[string]RefreshToken{},
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

func (db *DB) LoadDB() (DBStructure, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var dbStructure DBStructure
	dat, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}
	if err != nil {
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
	dbStructure, err := db.LoadDB()
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
	dbStruct, err := db.LoadDB()
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
	fmt.Println("DB authorization active")
	dbStruct, err := db.LoadDB()
	if err != nil {
		return UserDat{}, err
	}
	fmt.Println("DB loaded")

	// Normalize email inputs for safer comparison
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	// Iterate through the map to find the email
	for _, user := range dbStruct.Users {
		fmt.Printf("Checking user: %s\n", user.Email)
		if strings.ToLower(strings.TrimSpace(user.Email)) == normalizedEmail {
			fmt.Println("User found!")
			return user, nil
		}
	}
	fmt.Println("User not found!")
	return UserDat{}, errors.New("user not found")
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
