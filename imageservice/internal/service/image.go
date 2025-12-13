package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/domain"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/repository"
)

// TODO: error domain, logging

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

func (s *ImageService) UploadImage(ctx context.Context, imageMetadata *domain.ImageMetadata, object []byte) error {
	reader := bytes.NewReader(object)

	opts := minio.PutObjectOptions{
		ContentType: "image/jpeg",
	}
	err := s.movieImageRepo.Upload(ctx, imageMetadata.Bucket, imageMetadata.Name, reader, int64(len(object)), opts)
	if err != nil {
		return fmt.Errorf("error uploading image: %w", err)
	}

	return nil
}

func (s *ImageService) GetImage(ctx context.Context, imageMetadata *domain.ImageMetadata) (io.Reader, error) {
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
