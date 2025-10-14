package main

import (
	"database/sql"
	"log/slog"

	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
	"github.com/vladgrskkh/movie_recomendation_system/internal/mailer"
)

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer
}

func newApplication(cfg config, logger *slog.Logger, db *sql.DB, mailer mailer.Mailer) application {
	return application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer,
	}
}
