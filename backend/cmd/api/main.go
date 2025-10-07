package main

import (
	"database/sql"
	"flag"
	"log"
	"time"

	"github.com/vladgrskkh/movie_recomendation_system/internal/data"
)

const version = "1.0.0"

type application struct {
	config config
	logger *log.Logger
	db     data.MovieModel
	// mailer *mailer.Mailer
}

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

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://movie_user:password@localhost/movies?sslmode=disable", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "5m", "PostgreSQL max idle time")

	flag.Parse()

	logger := log.New(log.Writer(), "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	app := &application{
		config: cfg,
		logger: logger,
		db:     data.MovieModel{DB: db},
	}

	logger.Printf("Starting server on port %d in %s mode", cfg.port, cfg.env)
	if err := app.server(); err != nil {
		logger.Fatal(err)
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

// TO DO: add some handlers for different routes
// TO DO: think about the endpoints we need for the movie recommendation system
// TO DO: connect to the database
// TO DO: implement authentication and authorization
// TO DO: add logging and error handling
// TO DO: write tests for the handlers and other components
// TO DO: add mailer for user notifications and authentication/authorization
// TO DO: graceful shutdown and cleanup
// TO DO: rate limiter
// TO DO: CORS handling

// Tasks for today
// client -> reverse proxy (caddy) -> server -> postgres
// TO DO: health check endpoint
// JSON responses
// TO DO: environment variables for configuration
