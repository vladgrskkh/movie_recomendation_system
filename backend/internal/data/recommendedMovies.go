package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type recommendedMoviesModel struct {
	rdb *redis.Client
}

func (m recommendedMoviesModel) Get(userID int64) ([]*Movie, error) {
	js, err := m.rdb.Get(context.Background(), strconv.FormatInt(userID, 10)).Bytes()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return nil, ErrRecordNotFound
		default:
			return nil, fmt.Errorf("error getting recommended movies for redis: %w", err)
		}
	}

	var movies []*Movie
	if err := json.Unmarshal(js, &movies); err != nil {
		return nil, fmt.Errorf("error unmarshalling recommended movies from redis: %w", err)
	}

	return movies, nil
}

func (m recommendedMoviesModel) Set(userID int64, movies []*Movie) error {
	js, err := json.Marshal(movies)
	if err != nil {
		return fmt.Errorf("error marshalling recommended movies to redis: %w", err)
	}

	err = m.rdb.Set(context.Background(), strconv.FormatInt(userID, 10), js, 30*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("error setting recommended movies in redis: %w", err)
	}

	return nil
}
