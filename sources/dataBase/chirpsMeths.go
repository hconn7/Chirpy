package database

import "errors"

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
func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:       id,
		Body:     body,
		AuthorID: authorID,
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

func (db *DB) DeleteChirp(chirpID, userID int) error {
	dbStruct, err := db.LoadDB()
	if err != nil {
		return err
	}

	chirp, ok := dbStruct.Chirps[chirpID]
	if !ok {
		return errors.New("Chirp not found")
	}

	if chirp.AuthorID != userID {
		return errors.New("Chirp someone your own size!")
	}

	delete(dbStruct.Chirps, chirpID)

	if err := db.writeDB(dbStruct); err != nil {
		return err
	}

	return nil
}
func (db *DB) GetChirpByAuth(authorID int) ([]Chirp, error) {
	dbStruct, err := db.LoadDB()
	if err != nil {
		return nil, errors.New("Couldn't load db")
	}

	var authChirps []Chirp
	for _, chirp := range dbStruct.Chirps {
		if chirp.AuthorID == authorID {
			authChirps = append(authChirps, chirp)
		}
	}
	return authChirps, nil
}
