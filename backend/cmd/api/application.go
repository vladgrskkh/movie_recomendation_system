package main

import (
	"database/sql"
	"log/slog"
	"sync"

	"google.golang.org/grpc"

	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
	"github.com/vladgrskkh/movie_recomendation_system/internal/kafka"
)

type application struct {
	config   config
	logger   *slog.Logger
	models   data.Models
	grpcConn *grpc.ClientConn
	wg       sync.WaitGroup
	producer *kafka.Producer
}

func newApplication(cfg config, logger *slog.Logger, db *sql.DB, grpcConn *grpc.ClientConn, producer *kafka.Producer) application {
	return application{
		config:   cfg,
		logger:   logger,
		models:   data.NewModels(db),
		grpcConn: grpcConn,
		producer: producer,
	}
}
