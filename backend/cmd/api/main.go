package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
)

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
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "5m", "PostgreSQL max idle time")

	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, loggerOpts))

	ctx := context.Background()

	db, err := openDB(cfg)
	if err != nil {
		logger.Log(ctx, LevelFatal, err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("database connection pool established")

	app := newApplication(cfg, logger, db)

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
// TO DO: logger and error handling
// TO DO: context
// TO DO: some middleware
// TO DO: user auth? (mailer?)
// ::::::::::::::::::::::::::::::::

// TO DO: implement authentication and authorization
// TO DO: add logging and error handling
// TO DO: write tests for the handlers and other components
// TO DO: add mailer for user notifications and authentication/authorization
// TO DO: graceful shutdown and cleanup
// TO DO: rate limiter
// TO DO: CORS handling
