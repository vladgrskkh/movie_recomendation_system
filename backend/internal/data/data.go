package data

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type moviesInterface interface {
	Get(int64) (*Movie, error)
	Insert(*Movie) error
	Delete(int64) error
	Update(*Movie) error
	GetAll(string, []string, Filters) ([]*Movie, Metadata, error)
}

type usersInterface interface {
	Insert(*User) error
	GetByEmail(string) (*User, error)
	GetByID(int64) (*User, error)
	Update(*User) error
	GetForToken(string, string) (*User, error)
}

type tokensInterface interface {
	New(userID int64, ttl time.Duration, scope string) (*Token, error)
	Insert(token *Token) error
	DeleteAllForUser(scope string, userID int64) error
}

type Models struct {
	Movies moviesInterface
	Users  usersInterface
	Tokens tokensInterface
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: movieModel{DB: db},
		Users:  userModel{DB: db},
		Tokens: tokenModel{DB: db},
	}
}

// TODO: read about interface and how it should be for mocking dependency
