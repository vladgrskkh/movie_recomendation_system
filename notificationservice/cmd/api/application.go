package main

import (
	"log/slog"

	"github.com/vladgrskkh/movie_recomendation_system/notificationservice/internal/consumer"
	"github.com/vladgrskkh/movie_recomendation_system/notificationservice/internal/mailer"
)

type application struct {
	config          config
	logger          *slog.Logger
	mailer          *mailer.Mailer
	mailerConsumers []*consumer.Consumer
}

func newApplication(cfg config, logger *slog.Logger, mailer *mailer.Mailer) *application {
	return &application{
		config: cfg,
		logger: logger,
		mailer: mailer,
	}
}
