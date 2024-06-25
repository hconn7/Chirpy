package database

import (
	"encoding/json"
	"errors"
	"os"
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
	Subscription bool   `json:"is_chirpy_red"`
}

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
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
