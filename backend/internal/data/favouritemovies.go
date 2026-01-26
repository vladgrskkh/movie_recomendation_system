package data

import (
	"context"
	"time"
)

func (m movieModel) InsertFavourite(movieID, userID int64) error {
	query := `
		INSERT INTO movie_favourite (movie_id, user_id)
		VALUES $1, $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, movieID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (m movieModel) DeleteFavourite(movieID, userID int64) error {
	query := `
		DELETE FROM movie_favourite
		WHERE movie_id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, movieID, userID)
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
