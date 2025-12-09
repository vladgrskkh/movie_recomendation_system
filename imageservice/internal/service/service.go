package service

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/repository"
)

// TODO: error domain, logging
// TODO: minio opts

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

func (s *ImageService) UploadImage(ctx context.Context, bucketName string, objectName string, object []byte, opts minio.PutObjectOptions) error {
	reader := bytes.NewReader(object)

	err := s.movieImageRepo.Upload(ctx, bucketName, objectName, reader, int64(len(object)), opts)
	// err domain
	if err != nil {
		return err
	}

	return nil
}

func (s *ImageService) GetImage(ctx context.Context, bucketName string, objectName string, opts minio.GetObjectOptions) ([]byte, error) {
	object, err := s.movieImageRepo.Get(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, err
	}

	info, err := object.Stat()
	if err != nil {
		return nil, err
	}

	image := make([]byte, info.Size)

	_, err = object.Read(image)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (s *ImageService) DeleteImage(ctx context.Context, bucketName string, objectName string, opts minio.RemoveObjectOptions) error {
	err := s.movieImageRepo.Delete(ctx, bucketName, objectName, opts)
	if err != nil {
		return err
	}

	return nil
}
