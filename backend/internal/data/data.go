package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type moviesInterface interface {
	Get(id int64) (*Movie, error)
	GetByIDs(ids []int64) ([]*Movie, error)
	GetByTitle(title string) (int64, error)
	Insert(movie *Movie) error
	Delete(id int64) error
	Update(movie *Movie) error
	GetAll(title string, similarityThreshold float64, genres []string, fileters Filters) ([]*Movie, Metadata, error)
	GetPopular() ([]*Movie, error)
	GetNew(year int) ([]*Movie, error)
	InsertWatched(id int64, userID int64) error
	GetWatched(userID int64) ([]int64, error)
}

type recommenderMoviesInterface interface {
	Get(userID int64) ([]*Movie, error)
	Set(userID int64, movies []*Movie) error
}

type usersInterface interface {
	Insert(*User) error
	GetByEmail(string) (*User, error)
	GetByID(int64) (*User, error)
	Update(*User) error
	GetForToken(string, string) (*User, error)
	Delete(*User) error
}

type tokensInterface interface {
	New(userID int64, ttl time.Duration, scope string) (*Token, error)
	Insert(token *Token) error
	DeleteAllForUser(scope string, userID int64) error
}

type Models struct {
	Movies            moviesInterface
	RecommendedMovies recommenderMoviesInterface
	Users             usersInterface
	Tokens            tokensInterface
}

func NewModels(db *sql.DB, rdb *redis.Client) Models {
	return Models{
		Movies:            movieModel{DB: db},
		RecommendedMovies: recommendedMoviesModel{rdb: rdb},
		Users:             userModel{DB: db},
		Tokens:            tokenModel{DB: db},
	}
}

// TODO: read about interface and how it should be for mocking dependency
