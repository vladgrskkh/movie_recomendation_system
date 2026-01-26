package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/minio/minio-go/v7"
)

var (
	ErrNotFound     = errors.New("object not found")
	ErrNoSuchBucket = errors.New("bucket not exists")
)

type MovieImageRepo struct {
	logger  *slog.Logger
	storage *minio.Client
}

func NewMovieImageRepo(logger *slog.Logger, storage *minio.Client) *MovieImageRepo {
	return &MovieImageRepo{
		logger:  logger,
		storage: storage,
	}
}

func (r *MovieImageRepo) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	exists, err := r.storage.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("error checking if bucket exists: %w", err)
	}

	if exists {
		return nil
	}

	err = r.storage.MakeBucket(ctx, bucketName, opts)
	if err != nil {
		return fmt.Errorf("error creating bucket: %w", err)
	}

	return nil
}

func (r *MovieImageRepo) Upload(ctx context.Context, bucketName string, objectName string, object io.Reader, size int64, opts minio.PutObjectOptions) error {
	info, err := r.storage.PutObject(ctx, bucketName, objectName, object, size, opts)
	if err != nil {
		v, ok := err.(*minio.ErrorResponse)
		if ok {
			switch v.Code {
			case minio.NoSuchBucket:
				return ErrNoSuchBucket
			default:
				return err
			}
		}
		return fmt.Errorf("error uploading file: %w", err)
	}

	r.logger.Info("Successfully uploaded file", slog.String("objectName", objectName), slog.Int64("size", size), slog.Int64("size", info.Size))
	return nil
}

func (r *MovieImageRepo) Get(ctx context.Context, bucketName string, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	object, err := r.storage.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		v, ok := err.(*minio.ErrorResponse)
		if ok {
			switch v.Code {
			case minio.NoSuchKey:
				return nil, ErrNotFound
			case minio.NoSuchBucket:
				return nil, ErrNotFound
			default:
				return nil, err
			}
		}
		return nil, err
	}

	r.logger.Info("Successfully fetched file", slog.String("objectName", objectName), slog.String("bucket", bucketName))
	return object, nil
}

func (r *MovieImageRepo) Delete(ctx context.Context, bucketName string, objectName string, opts minio.RemoveObjectOptions) error {
	err := r.storage.RemoveObject(ctx, bucketName, objectName, opts)
	if err != nil {
		v, ok := err.(*minio.ErrorResponse)
		if ok {
			switch v.Code {
			case minio.NoSuchKey:
				return ErrNotFound
			case minio.NoSuchBucket:
				return ErrNotFound
			}
		}

		return err
	}

	r.logger.Info("Successfully deleted file", slog.String("objectName", objectName), slog.String("bucket", bucketName))
	return nil
}
