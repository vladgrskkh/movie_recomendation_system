package data

import (
	"database/sql"
)

type Movie struct {
	ID      int    `json:"id"`
	Year    int    `json:"year"`
	Title   string `json:"title"`
	Runtime int    `json:"runtime"`
	Genres  string `json:"genres"`
}

type MovieModel struct {
	DB *sql.DB
}

// client side api calling

func (m MovieModel) Get(id int64) (*Movie, error) {
	query := `
		SELECT id, year, title, runtime, genres
		FROM movies
		WHERE id = $1
	`

	var movie Movie

	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.Year,
		&movie.Title,
		&movie.Runtime,
		&movie.Genres,
	)
	if err != nil {
		return nil, err
	}

	return &movie, nil
}
