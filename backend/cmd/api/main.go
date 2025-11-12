package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/vladgrskkh/movie_recomendation_system/internal/kafka"
	"github.com/vladgrskkh/movie_recomendation_system/internal/mailer"
)

// @title Movie Recommendation System API
// @version 1.0.0
// @description REST API for recomending movies, managing users and authentication.
// @BasePath /v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Provide a Bearer token: "Bearer {token}"

var (
	buildTime string
	version   string
)

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
		mailerAPIKey string
		sender       string
	}
	limiter struct {
		rps    int
		enable bool
	}
	grpc struct {
		address string
	}
	jwt struct {
		secretKey      string
		secretKeyBytes []byte
	}
	kafka struct {
		address []string
		topic   string
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

	flag.StringVar(&cfg.smtp.mailerAPIKey, "smtp-mailer-api-key", "", "SMTP MailerSend API key")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "", "SMTP sender")

	flag.IntVar(&cfg.limiter.rps, "limiter-rps", 10, "Rate limiter maximum requests per second")
	flag.BoolVar(&cfg.limiter.enable, "limiter-enable", true, "Enable rate limiter")

	flag.StringVar(&cfg.grpc.address, "grpc-address", "", "gRPC server address")

	flag.StringVar(&cfg.jwt.secretKey, "jwt-secret", "", "Secret key for signing and verifying JWT tokens")

	flag.StringVar(&cfg.kafka.topic, "kafka-topic", "", "Kafka topic")
	flag.Func("kafka-address", "addresses for kafka brokers", func(s string) error {
		if s == "" {
			return fmt.Errorf("kafka-address flag cannot be empty")
		}

		cfg.kafka.address = strings.Split(s, ",")
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and quit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Build time:\t%s\nVersion:\t%s\n", buildTime, version)
		os.Exit(0)
	}

	// Convert JWT secret key to byte slice
	cfg.jwt.secretKeyBytes = []byte(cfg.jwt.secretKey)

	mailer := mailer.New(cfg.smtp.mailerAPIKey, cfg.smtp.sender)

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
		} else if e != nil {
			logger.Error(err.Error())
		}
	}()

	logger.Info("database connection pool established")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(cfg.grpc.address+":50051", opts...)
	if err != nil {
		logger.Log(ctx, LevelFatal, "cannot connect to gRPC server: "+err.Error())
		os.Exit(1)
	}

	defer func() {
		e := conn.Close()
		if err != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else if e != nil {
			logger.Error(err.Error())
		}
	}()

	logger.Info("gRPC connection established")

	p, err := kafka.NewProducer(cfg.kafka.address)
	if err != nil {
		logger.Log(ctx, LevelFatal, err.Error())
		os.Exit(1)
	}

	logger.Info("new kafka producer started")

	app := newApplication(cfg, logger, db, mailer, conn, p)

	logger.Info("starting server", slog.Int("port", cfg.port), slog.String("environment", cfg.env))
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
// TO DO: write tests for the handlers and other components (2 hours)
// ::::::::::::::::::::::::::::::::

// TO DO: write tests for the handlers and other components
// TO DO: think about how to serve images for movies
// TO DO: user profile handler
// TODO: add more metrics, grafana settings (best practice)
// TODO: add redis db for ip rate limmiter
// TODO: make use of makefile in cicd pipelines
// TODO: grafana storage persistence
// TODO: need to check if i may need more than one producer
