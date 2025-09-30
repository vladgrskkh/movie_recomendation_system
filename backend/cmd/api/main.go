package main

import (
	"flag"
	"log"
)

const version = "1.0.0"

type application struct {
	config config
	logger *log.Logger
	// db     *sql.DB
	// mailer *mailer.Mailer
}

type config struct {
	port int
	env  string
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	logger := log.New(log.Writer(), "", log.Ldate|log.Ltime)

	app := &application{
		config: cfg,
		logger: logger,
	}

	logger.Printf("Starting server on port %d in %s mode", cfg.port, cfg.env)
	if err := app.server(); err != nil {
		logger.Fatal(err)
	}
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
