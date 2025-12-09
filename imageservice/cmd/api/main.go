package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/cfg"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/handlers"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/repository"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))

	logger.Info("parsing .env into config")

	cfg, err := cfg.NewConfig()
	if err != nil {
		logger.Error("somethings wrong with .env", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// minio connection.
	logger.Info("connecting to minio server")
	client, err := minio.New(cfg.Minio.Endopint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""),
		Secure: false, // in the same container (mb tune this later cause its will run under caddy)
	})
	if err != nil {
		logger.Error("error connecting to minio server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	movieImageRepo := repository.NewMovieImageRepo(logger, client)
	// creating image bucket if not exists
	err = movieImageRepo.MakeBucket(context.TODO(), cfg.Minio.BucketName, minio.MakeBucketOptions{})
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	s := service.NewImageService(logger, movieImageRepo)

	// grpc server
	h := handlers.NewImageHandler(logger, s)

	logger.Info("starting grpc server")
	err = server(logger, cfg.Address, h)
	if err != nil {
		logger.Error("failed to start grpc server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// TODO: minio repo
// TODO: grpc contract
// TODO: think about how to serve images
