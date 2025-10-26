package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/vladgrskkh/movie_recomendation_system/internal/mailer"
)

// @title Movie Recommendation System API
// @version 1.0.0
// @description REST API for recomending movies, managing users and authentication.
// @BasePath /v1
// @schemes https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Provide a Bearer token: "Bearer {token}"

const version = "1.0.0"

const (
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
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "5m", "PostgreSQL max idle time")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailersend.net", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "MS_rGhIm9@test-r83ql3pqpxxgzw1j.mlsender.net", "SMTP sender")

	flag.Parse()

	mailer := mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender)

	logger := slog.New(slog.NewJSONHandler(os.Stderr, loggerOpts))

	ctx := context.Background()

	db, err := openDB(cfg)
	if err != nil {
		logger.Log(ctx, LevelFatal, err.Error())
		os.Exit(1)
	}

	defer func() {
		e := db.Close()
		if err != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else {
			err = e
		}
	}()

	logger.Info("database connection pool established")

	app := newApplication(cfg, logger, db, mailer)

	logger.Info("Starting server", slog.Int("port", cfg.port), slog.String("environment", cfg.env))
	if err := app.server(); err != nil {
		logger.Log(ctx, LevelFatal, err.Error())
		os.Exit(1)
	}
}

// openDB opens a database connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Task for today::::::::::::::::::
// ::::::::::::::::::::::::::::::::
// TO DO: metrics (prometheus, grafana, expvar etc) (idk 4 house mb)
// TO DO: makefile new rules (30 min)
// TO DO: write tests for the handlers and other components (2 hours)
// TO DO: routes groups (30 min)
// TO DO: rate limiter (mb use already implemented stuff from github )(1-2 hours)
// fix bug: mailer on vps dial i/o timeout
// ::::::::::::::::::::::::::::::::

// TO DO: write tests for the handlers and other components
// TO DO: rate limiter
// TO DO: routes groups
// TO DO: think about how to serve images for movies
// TO DO: metrics (prometheus, grafana, expvar etc)
// TO DO: python ml microservice (grpc)
// TO DO: reset password handler
// TO DO: user profile handler
// bug: mailer on vps dial i/o timeout
