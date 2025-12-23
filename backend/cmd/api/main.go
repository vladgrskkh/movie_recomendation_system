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

	"github.com/redis/go-redis/v9"
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
	redis struct {
		address  string
		password string
		username string
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

	flag.StringVar(&cfg.redis.address, "redis-address", "redis:6379", "Redis address")
	flag.StringVar(&cfg.redis.password, "redis-password", "default-password", "Redis password")
	flag.StringVar(&cfg.redis.username, "redis-username", "default-username", "Redis username")

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

	// TODO: helper func wrapper for this
	defer func() {
		e := db.Close()
		if err != nil && e != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else if e != nil {
			err = e
		}
	}()

	logger.Info("postgres database connection pool established")

	rdb, err := redisClient(cfg)
	if err != nil {
		logger.Log(ctx, LevelFatal, "cannot connect to redis:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() {
		e := rdb.Close()
		if err != nil && e != nil {
			err = fmt.Errorf("previous error: %w; close error: %w", err, e)
		} else if e != nil {
			err = e
		}
	}()

	logger.Info("redis connection pool established")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	connRecommender, err := grpc.NewClient(cfg.grpc.address+":50051", opts...)
	if err != nil {
		logger.Log(ctx, LevelFatal, "cannot connect to gRPC recommender server:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	clientRecommender := predict.NewRecommendationClient(connRecommender)
	// TODO: fetch from config
	connImage, err := grpc.NewClient("image-service:50052", opts...)
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

	app := newApplication(cfg, logger, db, rdb, clientRecommender, clientImage, p)

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

func redisClient(cfg config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.redis.address,
		Password: cfg.redis.password,
		DB:       0,
		Username: cfg.redis.username,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}

// Task for today::::::::::::::::::
// ::::::::::::::::::::::::::::::::
// TODO: write tests for the handlers and other components (2 hours)
// ::::::::::::::::::::::::::::::::

// TODO: write tests for the handlers and other components
// TODO: user profile handler
// TODO: add more metrics, grafana settings (best practice)
// TODO: make use of makefile in cicd pipelines
// TODO: need to check if i may need more than one producer (worker pool)
// TODO: prometheus work around duplicate metrics with tests
// TODO: fix bug with github tags in ci/cd pipeline
// TODO: ci/cd for image service
// TODO: handlers_test remade
// TODO: insert new model into predict service
// TODO: seed data in db and post images to image service for movies
// TODO: update ci/cd pipeline
// TODO: revision on uploading image
// TODO: 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/l43OTqhib8do1WwXig8JI8Y68Fd.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/hi1E20BEDuOAsEpJ4rLM09MLRfZ.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/5DGlfuO31upq3ieF5rHBVALz5kn.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/1fAIdDUVhQaqBr4lVPOeEiNPA3l.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/A3yr3FSiE6ECTQAPs5TFmHWKXBy.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/mO9uOoFhuFsq4iFCAdfRqndcvzW.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/w5ry8VK2mNbINQwC0ZBpCdBnaPx.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/kOqZpuN5Q3F1GU6SFJjWnYd2uBo.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/sUZfWQ8zlZx0OwOTmorDgttzIia.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/bziiIpZUOkhtZA2IMsurHywgc9Y.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/i5MaRm0FsJIpXVLAvrMsMt5YQ5p.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/5WzusZileLb3TLyMOYrwWrh4YXu.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/dDIAA48bsuJPZk1JIFIneNw1XwF.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/exk2xDI2YT5UHs5Bow79AVRGOnP.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/8Tnn2UOlecCKNLKG8ZK5Mu6bK94.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/84jdp4kDVL9jm5VmWyELd7K13jw.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/9brJrqCRm1kH0duQ4mclYblECrg.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/e0tZ5Kjg5mukFGcxEddCZymAxH4.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/yRKyJJYIzfeiVDHBe4LXguPQCvD.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/a3eYLIK5eYSHAJTtOy9PdciVjxx.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/k0oT1Vj9dlkwk5Vuuz7pvnG8iJ6.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/3v06AHzlKIsr9CV75m6oxJHTCpy.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/xDDu9iup7qobqSn6yYjsTNex1mt.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/pO5XR2R56RAbVjdks9gGGn0fbOa.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/838yNJnOk6iZj5YK1MoKiN2n1hw.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/aFt9A4TK8n869uBsOBJj5cfZRDi.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/jqu2wqiFKrmhbF6fLqwi3zRUDz7.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/feiECea2yhyiNRAfvqXzbQFgR1W.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/ob8TdzmXl6ITljpdKJDVZZpwLbw.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/8edMloIwppGDsJ4tqqhDfH9IOSN.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/eKtkT0INpwRwJUGSxxQQ1ZckHHM.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/gd5EoAU4MM57sW3vlWxJ0NMM8cV.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/jiGn1FXSQVfN2nhix0tQtmoGWcF.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/a5kImJOOgKOiCNHvYwTJPnJMzQr.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/1ohDQEb2UVWPyCJcvxFmP2hJFLN.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/dolTehtRFO6B6yVu21yU7TBShot.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/5IYyJetEctAypFYIffqx55PXTPT.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/vV40BJbN1NVYkOXVdkQt4qi1PmQ.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/8H67aqYnvwA9C36tX3Awh8vimJj.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/rlmUnv02eyEzJWaAbl8yRbhZDep.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/rGaXfUvsnIK32RikoCnYKoYsQNc.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/9t6YnN0DsIK530RWsWjVdgpcKPe.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/zjRkl5PZvFujbRndrNbf4ZKbeNC.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/tSj2g6VdJ2UXeWk9wRa19SsxRFg.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/uUXLq7fEG3hI46ZFMZzgHj11S6S.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/bXPKTWAuRO8Mu5rLCORFagX0YZ1.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w1280/sZfK9Gy0mHbGjPte0lgXTprmIsI.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/rESAN6IaAkoblJPW4pOgjOqFNhA.jpg": unexpected EOF
// 2025/12/21 22:05:54 ERROR Get "https://image.tmdb.org/t/p/w780/ydPsTcrbPKdIN0OJwZ5n4xDYlf.jpg": unexpected EOF
