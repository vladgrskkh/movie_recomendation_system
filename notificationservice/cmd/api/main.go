package main

import (
	"log/slog"
	"os"

	"github.com/vladgrskkh/movie_recomendation_system/notificationservice/config"
	"github.com/vladgrskkh/movie_recomendation_system/notificationservice/internal/mailer"
)

const (
	// logger levels
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

var LevelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

var loggerOpts = &slog.HandlerOptions{
	Level: LevelTrace,

	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.LevelKey {
			level := a.Value.Any().(slog.Level)
			levelLabel, exists := LevelNames[level]
			if !exists {
				levelLabel = level.String()
			}

			a.Value = slog.StringValue(levelLabel)
		}
		return a
	},
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, loggerOpts))

	cfg, err := config.New()
	if err != nil {
		logger.Error("Failed to create config file", slog.String("error", err.Error()))
		os.Exit(1)
	}

	mailer := mailer.NewMailer(cfg.Mailer.MailerAPIKey, cfg.Mailer.Sender)

	app := newApplication(cfg, logger, &mailer)

	app.logger.Info("Starting notification server")

	err = app.server()
	if err != nil {
		app.logger.Error("Something went wrong while shuttind down server", slog.String("error", err.Error()))
	}
}

// TODO: use mongodb to store details about message (think about why i may need this etc)
// TODO: add logic for push notification (firebase)
// TODO: check best practice for shutting down kafka consumers
// TODO: wrap errors and change log messages
// TODO: delete first docker containter
// TODO: cannot run notification containre (cannot find .server notification-service-1  | exec ./server: no such file or directory)
// TODO: pass .env to prod with github actions
// TODO: push notification with firebase
