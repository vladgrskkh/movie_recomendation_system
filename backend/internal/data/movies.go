package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Movie struct {
	ID           int64     `json:"id" example:"1"`
	CreatedAt    time.Time `json:"-"`
	Title        string    `json:"title" example:"The Shawshank Redemption"`
	Overview     string    `json:"overview" example:"Two imprisoned men bond over..."`
	Year         int32     `json:"year" example:"1994"`
	ReleaseDate  string    `json:"release_date" example:"1994-09-23"`
	Runtime      int32     `json:"runtime,omitempty" example:"142"`
	Genres       []string  `json:"genres" example:"Drama,Crime"`
	GenreIDs     []int32   `json:"-"`
	VoteAverage  float32   `json:"vote_average,omitempty" example:"8.7"`
	VoteCount    int32     `json:"vote_count,omitempty" example:"100"`
	PosterPath   string    `json:"poster_path,omitempty"`
	BackdropPath string    `json:"backdrop_path,omitempty"`
	Version      int32     `json:"version" example:"1"`
}

type movieModel struct {
	DB *sql.DB
}

func (m movieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT m.id, m.created_at, m.title, m.overview, m.release_date, m.year, m.runtime, m.vote_average, m.vote_count, m.poster_path, m.backdrop_path, m.version,
		ARRAY_AGG(g.name ORDER BY g.name) AS genres
		FROM movies
		LEFT JOIN movie_genres mg ON mg.movie_id = m.id
		LEFT JOIN genres g ON g.id = mg.genre_id
		WHERE m.id = $1
	`

	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Overview,
		&movie.ReleaseDate,
		&movie.Year,
		&movie.Runtime,
		&movie.VoteAverage,
		&movie.VoteCount,
		&movie.PosterPath,
		&movie.BackdropPath,
		&movie.Version,
		pq.Array(&movie.Genres),
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// TODO: use transaction here
func (m movieModel) Insert(movie *Movie) error {
	queryMovieGenres := `
		INSERT INTO movie_genres (movie_id, genre_id)
		SELECT $1, g.id
		FROM genres g
		WHERE g.name = ANY($2);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, queryMovieGenres, movie.ID, movie.Genres)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO movies (id, title, year, runtime)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, version
	`
	err = m.DB.QueryRowContext(ctx, query,
		movie.ID,
		movie.Title,
		movie.Year,
		movie.Runtime,
	).Scan(&movie.CreatedAt, &movie.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m movieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE FROM movies
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// TODO: use transaction here
func (m movieModel) Update(movie *Movie) error {
	query := `
	UPDATE movies
	SET title = $1, year = $2, runtime = $3, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, movie.Title, movie.Year, movie.Runtime, movie.ID, movie.Version).Scan(
		&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	queryMovieGenresDelete := `
		DELETE FROM movie_genres
		WHERE movie_id = $1;
	`

	_, err = m.DB.ExecContext(ctx, queryMovieGenresDelete, movie.ID)
	if err != nil {
		return err
	}

	queryMovieGenresInsert := `
		INSERT INTO movie_genres (movie_id, genre_id)
		SELECT $1, g.id
		FROM genres g
		WHERE g.name = ANY($2);
	`

	_, err = m.DB.ExecContext(ctx, queryMovieGenresInsert, movie.ID, movie.Genres)
	if err != nil {
		return err
	}

	return nil
}

func (m movieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	// TODO: need to test this query
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), m.id, m.created_at, m.title, m.year, m.runtime, ARRAY_AGG(g.name ORDER BY g.name) AS genres, m.version
	FROM movies m
	LEFT JOIN movie_genres mg ON mg.movie_id = m.id
	LEFT JOIN genres g ON g.id = mg.genre_id
	WHERE (to_tsvector('simple', m.title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND ($2 = '{}' OR EXISTS (
        SELECT 1
        FROM movie_genres mg2
        JOIN genres g2 ON g2.id = mg2.genre_id
        WHERE mg2.movie_id = m.id AND g2.name = ANY($2)
    ))
	GROUP BY m.id, m.created_at, m.title, m.year, m.runtime, m.version
	ORDER BY %s %s, id ASC
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rowsMovies, err := m.DB.QueryContext(ctx, query, title, pq.Array(genres), filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	defer func() {
		e := rowsMovies.Close()
		if err != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else {
			err = e
		}
	}()

	totalRecords := 0
	var movies []*Movie

	for rowsMovies.Next() {
		var movie Movie

		err := rowsMovies.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		movies = append(movies, &movie)
	}

	if err = rowsMovies.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}

func (m movieModel) GetPopular() ([]*Movie, error) {
	query := `
		SELECT id, title, year, poster_path, backdrop_path, version FROM movies
		JOIN popular_movies ON movies.id = popular_movies.id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer func() {
		e := rows.Close()
		if err != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else {
			err = e
		}
	}()

	var movies []*Movie

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			movie.ID,
			movie.Title,
			movie.Year,
			movie.PosterPath,
			movie.BackdropPath,
			movie.Version,
		)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m movieModel) GetNew(year int) ([]*Movie, error) {
	query := `
		SELECT id, title, year, poster_path, backdrop_path, version FROM movies
		WHERE year = $1
		LIMIT 20
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}

	defer func() {
		e := rows.Close()
		if err != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else {
			err = e
		}
	}()

	var movies []*Movie

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			movie.ID,
			movie.Title,
			movie.Year,
			movie.PosterPath,
			movie.BackdropPath,
			movie.Version,
		)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m movieModel) GetRecommended() ([]*Movie, error) {
	query := `
		SELECT id, title, year, poster_path, backdrop_path, version FROM movies
		JOIN recommended_movies ON movies.id = recommended_movies.id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer func() {
		e := rows.Close()
		if err != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else {
			err = e
		}
	}()

	var movies []*Movie

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			movie.ID,
			movie.Title,
			movie.Year,
			movie.PosterPath,
			movie.BackdropPath,
			movie.Version,
		)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}
