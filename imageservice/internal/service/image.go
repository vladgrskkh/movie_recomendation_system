package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/domain"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/repository"
)

var ErrInvalidImageCredentials = errors.New("invalid image credentials")

type ImageService struct {
	logger         *slog.Logger
	movieImageRepo *repository.MovieImageRepo
}

func NewImageService(logger *slog.Logger, movieImageRepo *repository.MovieImageRepo) *ImageService {
	return &ImageService{
		logger:         logger,
		movieImageRepo: movieImageRepo,
	}
}

func (s *ImageService) UploadImage(ctx context.Context, imageMetadata *domain.ImageMetadata, pr io.Reader) error {
	opts := minio.PutObjectOptions{
		ContentType: "image/jpeg",
	}
	s.logger.Debug("Starting image upload", slog.String("bucket", imageMetadata.Bucket), slog.String("filename", imageMetadata.Name), slog.Int64("size", imageMetadata.Size))

	err := s.movieImageRepo.Upload(ctx, imageMetadata.Bucket, imageMetadata.Name, pr, imageMetadata.Size, opts)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNoSuchBucket):
			return fmt.Errorf("error uploading image: %w", err)
		default:
			return err
		}
	}

	return nil
}

func (s *ImageService) GetImage(ctx context.Context, imageMetadata *domain.ImageMetadata) (io.Reader, error) {
	s.logger.Debug("Starting image fetch", slog.String("bucket", imageMetadata.Bucket), slog.String("filename", imageMetadata.Name), slog.Int64("size", imageMetadata.Size))

	object, err := s.movieImageRepo.Get(ctx, imageMetadata.Bucket, imageMetadata.Name, minio.GetObjectOptions{})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, domain.ErrorImageNotFound
		}

		return nil, err
	}

	return object, nil
}

func (s *ImageService) DeleteImage(ctx context.Context, imageMetadata *domain.ImageMetadata) error {
	s.logger.Debug("Starting image delete", slog.String("bucket", imageMetadata.Bucket), slog.String("filename", imageMetadata.Name), slog.Int64("size", imageMetadata.Size))

	err := s.movieImageRepo.Delete(ctx, imageMetadata.Bucket, imageMetadata.Name, minio.RemoveObjectOptions{})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return domain.ErrorImageNotFound
		}

		return err
	}

	return nil
}
