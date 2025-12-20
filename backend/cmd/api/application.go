package main

import (
	"database/sql"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/vladgrskkh/movie-recommender-contracts/v1/imageservice"
	"github.com/vladgrskkh/movie-recommender-contracts/v1/predict"

	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
	"github.com/vladgrskkh/movie_recomendation_system/internal/kafka"
)

type application struct {
	config        config
	logger        *slog.Logger
	models        data.Models
	predictClient predict.RecommendationClient
	imageClient   imageservice.ImageClient
	wg            sync.WaitGroup
	producer      *kafka.Producer
}

func newApplication(cfg config, logger *slog.Logger, db *sql.DB, rdb *redis.Client, predictClient predict.RecommendationClient, imageClient imageservice.ImageClient, producer *kafka.Producer) application {
	return application{
		config:        cfg,
		logger:        logger,
		models:        data.NewModels(db, rdb),
		predictClient: predictClient,
		imageClient:   imageClient,
		producer:      producer,
	}
}
