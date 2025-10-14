package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type movieInterface interface {
	Get(int64) (*Movie, error)
	Insert(*Movie) error
	Delete(int64) error
	Update(*Movie) error
}

type userInterface interface {
	Insert(*User) error
	GetByEmail(string) (*User, error)
	// Update()
}

type Models struct {
	Movies movieInterface
	Users  userInterface
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: movieModel{DB: db},
		Users:  userModel{DB: db},
	}
}
