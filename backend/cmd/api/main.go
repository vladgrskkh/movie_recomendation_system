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

	"github.com/vladgrskkh/movie-recommender-contracts/v1/imageservice"
	"github.com/vladgrskkh/movie-recommender-contracts/v1/predict"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/vladgrskkh/movie_recomendation_system/internal/kafka"
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
		address      []string
		topic        string
		passwordSSL  string
		username     string
		passwordUser string
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

	flag.IntVar(&cfg.limiter.rps, "limiter-rps", 10, "Rate limiter maximum requests per second")
	flag.BoolVar(&cfg.limiter.enable, "limiter-enable", true, "Enable rate limiter")

	flag.StringVar(&cfg.grpc.address, "grpc-address", "", "gRPC server address")

	flag.StringVar(&cfg.jwt.secretKey, "jwt-secret", "", "Secret key for signing and verifying JWT tokens")

	flag.StringVar(&cfg.kafka.topic, "kafka-topic", "", "Kafka topic")
	flag.StringVar(&cfg.kafka.passwordSSL, "kafka-password-ssl", "", "Kafka SSL password")
	flag.StringVar(&cfg.kafka.username, "kafka-username", "", "Kafka username for user SASL_SSL")
	flag.StringVar(&cfg.kafka.passwordUser, "kafka-password-user", "", "Kafka password for user SASL_SSL")
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

	logger := slog.New(slog.NewJSONHandler(os.Stderr, loggerOpts))

	ctx := context.Background()

	db, err := openDB(cfg)
	if err != nil {
		logger.Log(ctx, LevelFatal, "cannot connect to database:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() {
		e := db.Close()
		if err != nil && e != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else if e != nil {
			err = e
		}
	}()

	logger.Info("database connection pool established")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	connRecommender, err := grpc.NewClient(cfg.grpc.address+":50051", opts...)
	if err != nil {
		logger.Log(ctx, LevelFatal, "cannot connect to gRPC recommender server:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	clientRecommender := predict.NewRecommendationClient(connRecommender)

	connImage, err := grpc.NewClient(cfg.grpc.address+":50052", opts...)
	if err != nil {
		logger.Log(ctx, LevelFatal, "cannot connect to gRPC image server:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	clientImage := imageservice.NewImageClient(connImage)

	defer func() {
		e := connRecommender.Close()
		if err != nil && e != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else if e != nil {
			err = e
		}
	}()

	defer func() {
		e := connImage.Close()
		if err != nil && e != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else if e != nil {
			err = e
		}
	}()

	logger.Info("gRPC connections established")

	p, err := kafka.NewProducer(cfg.kafka.address, cfg.kafka.passwordSSL, cfg.kafka.username, cfg.kafka.passwordUser)
	if err != nil {
		logger.Log(ctx, LevelFatal, "cannot connect to kafka brokers", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("new kafka producer started")

	app := newApplication(cfg, logger, db, clientRecommender, clientImage, p)

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
// TODO: write tests for the handlers and other components (2 hours)
// TODO: add movies/new movies/recommended movies/popular endopints
// ::::::::::::::::::::::::::::::::

// TODO: write tests for the handlers and other components
// TODO: think about how to serve images for movies
// TODO: user profile handler
// TODO: add more metrics, grafana settings (best practice)
// TODO: add redis db for ip rate limmiter
// TODO: make use of makefile in cicd pipelines
// TODO: grafana storage persistence
// TODO: need to check if i may need more than one producer (worker pool)
// TODO: mb pass app to helper test methods instead of creating a new app(if tests is slow)
// TODO: prometheus work around duplicate metrics with tests
// TODO: mb separate services or change ci/cd pipeline
// TODO: fix bug with github tags in ci/cd pipeline
// TODO: add email input for activating user (also need to create separate table for activation/reset tokens)
