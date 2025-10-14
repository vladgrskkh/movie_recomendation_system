package main

import (
	"database/sql"
	"log/slog"

	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
)

type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func newApplication(cfg config, logger *slog.Logger, db *sql.DB) application {
	return application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}
}
