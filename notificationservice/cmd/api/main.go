package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"

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

var (
	loggerOpts = &slog.HandlerOptions{
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
)

type config struct {
	Address []string `toml:"address"`
	Mailer  struct {
		MailerAPIKey string `toml:"mailer_API_key"`
		Sender       string `toml:"sender"`
	} `toml:"mailer"`
	ConsumerMailer struct {
		ConsumerGroup string `toml:"consumer_group"`
		ConsumerCount int    `toml:"consumer_count"`
		Topic         string `toml:"topic"`
	} `toml:"consumer_mailer"`
}

func main() {
	var cfg config

	logger := slog.New(slog.NewJSONHandler(os.Stderr, loggerOpts))

	ctx := context.Background()

	metadata, err := toml.DecodeFile("config.toml", &cfg)
	if err != nil {
		logger.Log(ctx, LevelFatal, "error loading configuration", err.Error())
		os.Exit(1)
	}

	if len(metadata.Undecoded()) > 0 {
		logger.Log(ctx, LevelFatal, fmt.Sprintf("unknown configuration keys: %v", metadata.Undecoded()))
		os.Exit(1)
	}

	mailer := mailer.NewMailer(cfg.Mailer.MailerAPIKey, cfg.Mailer.Sender)

	app := newApplication(cfg, logger, &mailer)

	app.logger.Info("Starting notification server")

	err = app.server()
	if err != nil {
		app.logger.Log(ctx, LevelFatal, err.Error())
		os.Exit(1)
	}
}

// TODO: use mongodb to store details about message (think about why i may need this etc)
// TODO: add logic for push notification (firebase)
// TODO: check best practice for shutting down kafka consumers
